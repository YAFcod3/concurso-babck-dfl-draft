package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StatisticsResponse struct {
	TotalTransactions              int                `json:"totalTransactions"`
	TransactionsByType             map[string]int     `json:"transactionsByType"`
	TotalAmountConvertedByCurrency map[string]float64 `json:"totalAmountConvertedByCurrency"`
	FailedTransactionsInLast30Days int                `json:"failedTransactionsInLast30Days"`
	AverageAmountByTransactionType map[string]float64 `json:"averageAmountByTransactionType"`
}

func GetStatistics(c *fiber.Ctx, mongoClient *mongo.Client) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE_FORMAT",
				"message": "Start date is invalid. Use RFC3339 format.",
			})
		}
	} else {
		start = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	if endDate != "" {
		end, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE_FORMAT",
				"message": "End date is invalid. Use RFC3339 format.",
			})
		}
	} else {
		end = time.Now()
	}

	collection := mongoClient.Database("currencyMongoDb").Collection("transactions")

	pipeline := mongo.Pipeline{
		{{
			"$match", bson.M{"created_at": bson.M{"$gte": start, "$lt": end.Add(24 * time.Hour)}},
		}},
		{{
			"$lookup", bson.M{
				"from":         "transaction_types",
				"localField":   "transaction_type_id",
				"foreignField": "_id",
				"as":           "type_info",
			},
		}},
		{{
			"$unwind", "$type_info",
		}},

		{{
			"$group", bson.M{
				"_id": bson.M{
					"type_name": "$type_info.name",
					"currency":  "$to_currency",
				},
				"count":          bson.M{"$sum": 1},
				"totalConverted": bson.M{"$sum": "$amount_converted"},
				"totalAmount":    bson.M{"$sum": "$amount"},
			},
		}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate transactions"})
	}
	defer cursor.Close(context.Background())

	transactionsByType := make(map[string]int)
	totalAmountConvertedByCurrency := make(map[string]float64)
	totalAmountByType := make(map[string]float64)
	countByType := make(map[string]int)

	totalTransactions := 0
	for cursor.Next(context.Background()) {
		var result struct {
			ID struct {
				TypeName string `bson:"type_name"`
				Currency string `bson:"currency"`
			} `bson:"_id"`
			Count          int     `bson:"count"`
			TotalConverted float64 `bson:"totalConverted"`
			TotalAmount    float64 `bson:"totalAmount"`
		}

		if err := cursor.Decode(&result); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode result"})
		}

		transactionsByType[result.ID.TypeName] += result.Count
		totalAmountConvertedByCurrency[result.ID.Currency] += result.TotalConverted
		totalAmountByType[result.ID.TypeName] += result.TotalAmount
		countByType[result.ID.TypeName] += result.Count
		totalTransactions += result.Count
	}

	averageAmountByTransactionType := make(map[string]float64)
	for typeName, totalAmount := range totalAmountByType {
		count := countByType[typeName]
		if count > 0 {
			averageAmountByTransactionType[typeName] = totalAmount / float64(count)
		}
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	failedPipeline := mongo.Pipeline{
		{{
			"$match", bson.M{
				"created_at": bson.M{"$gte": thirtyDaysAgo},
				"status":     "failed",
			},
		}},
		{{
			"$count", "count",
		}},
	}

	failedCursor, err := collection.Aggregate(context.Background(), failedPipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count failed transactions"})
	}
	defer failedCursor.Close(context.Background())

	var failedResult struct {
		Count int `bson:"count"`
	}

	if failedCursor.Next(context.Background()) {
		if err := failedCursor.Decode(&failedResult); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode failed transactions count"})
		}
	}

	response := StatisticsResponse{
		TotalTransactions:              totalTransactions,
		TransactionsByType:             transactionsByType,
		TotalAmountConvertedByCurrency: totalAmountConvertedByCurrency,
		FailedTransactionsInLast30Days: failedResult.Count,
		AverageAmountByTransactionType: averageAmountByTransactionType,
	}

	return c.JSON(response)
}
