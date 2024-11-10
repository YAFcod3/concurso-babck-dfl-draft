package handlers

import (
	"exchange-rate/repository"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionResponse struct {
	TransactionId   string  `json:"transactionId"`
	TransactionCode string  `json:"transactionCode"`
	FromCurrency    string  `json:"fromCurrency"`
	ToCurrency      string  `json:"toCurrency"`
	Amount          float64 `json:"amount"`
	AmountConverted float64 `json:"amountConverted"`
	ExchangeRate    float64 `json:"exchangeRate"`
	TransactionType string  `json:"transactionType"`
	CreatedAt       string  `json:"createdAt"`
	UserId          string  `json:"userId"`
}

type TransactionsHistoryResponse struct {
	Total int                   `json:"total"`
	Data  []TransactionResponse `json:"data"`
}

func GetTransactions(c *fiber.Ctx, mongoClient *mongo.Client) error {
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	transactionType := c.Query("transactionType")

	// Convertir fechas a tipo time.Time
	var startDate, endDate time.Time
	var err error
	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid startDate format",
			})
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid endDate format",
			})
		}
	}

	filter := bson.M{
		"status": bson.M{"$eq": "successful"},
	}

	if !startDate.IsZero() && !endDate.IsZero() {
		filter["created_at"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	} else {
		if !startDate.IsZero() {
			filter["created_at"] = bson.M{"$gte": startDate}
		}
		if !endDate.IsZero() {
			filter["created_at"] = bson.M{"$lte": endDate}
		}
	}

	if transactionType != "" {
		filter["type_info.name"] = transactionType
	}

	database := mongoClient.Database("currencyMongoDb")

	// Obtener transacciones con el filtro especificado
	transactions, err := repository.GetFilteredTransactions(database, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Imprimir transacciones para depuración
	for _, transaction := range transactions {
		fmt.Printf("Transaction ID: %s, Type Name: %s\n", transaction.TransactionID.Hex(), transaction.TransactionTypeName)
	}

	// Construir la respuesta con el tipo de dato TransactionResponse
	response := TransactionsHistoryResponse{
		Total: len(transactions),
		Data:  []TransactionResponse{},
	}
	for _, transaction := range transactions {
		response.Data = append(response.Data, TransactionResponse{
			TransactionId:   transaction.TransactionID.Hex(),
			TransactionCode: transaction.TransactionCode,
			FromCurrency:    transaction.FromCurrency,
			ToCurrency:      transaction.ToCurrency,
			Amount:          transaction.Amount,
			AmountConverted: transaction.AmountConverted,
			ExchangeRate:    transaction.ExchangeRate,
			TransactionType: transaction.TransactionTypeName,
			CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
			UserId:          transaction.UserID,
		})
	}

	return c.JSON(response)
}
