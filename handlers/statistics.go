package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StatisticsResponse struct {
	TotalTransactions              int                `json:"totalTransactions"`
	TransactionsByType             map[string]int     `json:"transactionsByType"`
	TotalAmountConvertedByCurrency map[string]float64 `json:"totalAmountConvertedByCurrency"`
	TotalAmountByTransactionType   map[string]float64 `json:"totalAmountByTransactionType"`
	AverageAmountByTransactionType map[string]float64 `json:"averageAmountByTransactionType"`
}

func GetStatistics(c *fiber.Ctx, mongoClient *mongo.Client) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	fmt.Println("startDate : ", startDate)
	fmt.Println("endDate : ", endDate)

	// Parsear fechas
	start, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE_FORMAT",
			"message": "Start date is invalid. Use RFC3339 format.",
		})
	}

	end, err := time.Parse(time.RFC3339, endDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE_FORMAT",
			"message": "End date is invalid. Use RFC3339 format.",
		})
	}

	collection := mongoClient.Database("currencyMongoDb").Collection("transactions")

	// Pipeline de agregación
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"created_at": bson.M{"$gte": start, "$lt": end.Add(24 * time.Hour)}}}},
		{{"$group", bson.M{
			"_id":            "$transaction_type",
			"totalCount":     bson.M{"$sum": 1},
			"totalAmount":    bson.M{"$sum": "$amount"},
			"totalConverted": bson.M{"$sum": "$amount_converted"},
		}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate transactions"})
	}
	defer cursor.Close(context.Background())

	transactionsByType := make(map[string]int)
	totalAmountByTransactionType := make(map[string]float64)
	totalAmountConvertedByCurrency := make(map[string]float64)

	for cursor.Next(context.Background()) {
		var result struct {
			ID             string  `bson:"_id"`
			TotalCount     int     `bson:"totalCount"`
			TotalAmount    float64 `bson:"totalAmount"`
			TotalConverted float64 `bson:"totalConverted"`
		}

		if err := cursor.Decode(&result); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode result"})
		}

		transactionsByType[result.ID] = result.TotalCount
		totalAmountByTransactionType[result.ID] = result.TotalAmount
		totalAmountConvertedByCurrency[result.ID] = result.TotalConverted
	}

	// Calcular promedios
	averageAmountByTransactionType := make(map[string]float64)
	for tType, totalAmount := range totalAmountByTransactionType {
		count := transactionsByType[tType]
		if count > 0 {
			averageAmountByTransactionType[tType] = totalAmount / float64(count)
		}
	}

	response := StatisticsResponse{
		TotalTransactions:              sumValues(transactionsByType),
		TransactionsByType:             transactionsByType,
		TotalAmountConvertedByCurrency: totalAmountConvertedByCurrency,
		TotalAmountByTransactionType:   totalAmountByTransactionType,
		AverageAmountByTransactionType: averageAmountByTransactionType,
	}

	return c.JSON(response)
}

func sumValues(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

// GetStatistics maneja la obtención de estadísticas de transacciones
// func GetStatistics(c *fiber.Ctx, mongoClient *mongo.Client) error {
// 	startDate := c.Query("startDate")
// 	endDate := c.Query("endDate")

// 	// Parsear fechas
// 	start, err := time.Parse(time.RFC3339, startDate)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"code":    "INVALID_DATE_FORMAT",
// 			"message": "Start date is invalid. Use RFC3339 format.",
// 		})
// 	}

// 	end, err := time.Parse(time.RFC3339, endDate)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"code":    "INVALID_DATE_FORMAT",
// 			"message": "End date is invalid. Use RFC3339 format.",
// 		})
// 	}

// 	collection := mongoClient.Database("currencyMongoDb").Collection("transactions")

// 	// Filtrar transacciones por rango de fechas
// 	filter := bson.M{
// 		"created_at": bson.M{
// 			"$gte": start,
// 			"$lt":  end.Add(24 * time.Hour), // Para incluir todo el día de `endDate`
// 		},
// 	}

// 	// Contar total de transacciones
// 	totalTransactions, err := collection.CountDocuments(context.Background(), filter)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count transactions"})
// 	}

// 	// Obtener transacciones para calcular estadísticas
// 	cursor, err := collection.Find(context.Background(), filter)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch transactions"})
// 	}
// 	defer cursor.Close(context.Background())

// 	// Inicializar estructuras para estadísticas
// 	transactionsByType := make(map[string]int)
// 	totalAmountConvertedByCurrency := make(map[string]float64)
// 	totalAmountByTransactionType := make(map[string]float64)

// 	// Recorrer transacciones y acumular datos
// 	for cursor.Next(context.Background()) {
// 		var transaction Transaction
// 		if err := cursor.Decode(&transaction); err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode transaction"})
// 		}

// 		// Contar transacciones por tipo
// 		transactionsByType[transaction.TransactionType]++

// 		// Acumular total por moneda convertida
// 		totalAmountConvertedByCurrency[transaction.ToCurrency] += transaction.AmountConverted

// 		// Acumular total por tipo de transacción
// 		totalAmountByTransactionType[transaction.TransactionType] += transaction.Amount
// 	}

// 	// Calcular promedios
// 	averageAmountByTransactionType := make(map[string]float64)
// 	for tType, totalAmount := range totalAmountByTransactionType {
// 		count := transactionsByType[tType]
// 		if count > 0 {
// 			averageAmountByTransactionType[tType] = totalAmount / float64(count)
// 		}
// 	}

// 	// Crear respuesta
// 	response := StatisticsResponse{
// 		TotalTransactions:              int(totalTransactions),
// 		TransactionsByType:             transactionsByType,
// 		TotalAmountConvertedByCurrency: totalAmountConvertedByCurrency,
// 		TotalAmountByTransactionType:   totalAmountByTransactionType,
// 		AverageAmountByTransactionType: averageAmountByTransactionType,
// 	}

// 	return c.JSON(response)
// }
