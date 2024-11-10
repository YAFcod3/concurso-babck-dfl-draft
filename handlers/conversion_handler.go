package handlers

import (
	"context"
	"exchange-rate/models"
	"exchange-rate/utils/generate_transaction_code"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	StatusSuccessful models.TransactionStatus = "successful"
	StatusFailed     models.TransactionStatus = "failed"
	StatusPending    models.TransactionStatus = "pending"
)

func IsValidTransactionStatus(status models.TransactionStatus) bool {
	switch status {
	case StatusSuccessful, StatusFailed, StatusPending:
		return true
	default:
		return false
	}
}

type ConversionRequest struct {
	FromCurrency    string  `json:"fromCurrency"`
	ToCurrency      string  `json:"toCurrency"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transactionType"`
}

func ConvertCurrency(c *fiber.Ctx, mongoClient *mongo.Client, redisClient *redis.Client, codeGen *generate_transaction_code.CodeGenerator) error {
	var req ConversionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request",
		})
	}

	if req.FromCurrency == "" || req.ToCurrency == "" || req.TransactionType == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "All fields must be filled in and amount must be greater than zero",
		})
	}

	userId := c.Locals("userId").(string)

	transactionTypeID, err := primitive.ObjectIDFromHex(req.TransactionType)
	if err != nil {
		return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Invalid transaction type ID format")
	}

	collection := mongoClient.Database("currencyMongoDb").Collection("transaction_types")
	var transactionType struct {
		Name string `bson:"name"`
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": transactionTypeID}).Decode(&transactionType)
	if err != nil {
		return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Transaction type not found")
	}

	fromRateFloat := 1.0
	if req.FromCurrency != "USD" {
		fromRate, err := redisClient.HGet(context.Background(), "exchange_rates", req.FromCurrency).Result()
		if err != nil {
			return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "From currency not found")
		}
		fromRateFloat, err = strconv.ParseFloat(fromRate, 64)
		if err != nil {
			return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Invalid from currency rate")
		}
	}

	toRateFloat := 1.0
	if req.ToCurrency != "USD" {
		toRate, err := redisClient.HGet(context.Background(), "exchange_rates", req.ToCurrency).Result()
		if err != nil {
			return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "To currency not found")
		}
		toRateFloat, err = strconv.ParseFloat(toRate, 64)
		if err != nil {
			return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Invalid to currency rate")
		}
	}

	transactionCode, err := codeGen.GenerateCode()
	if err != nil {
		return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Failed to generate transaction code")
	}

	convertedAmount := (req.Amount / fromRateFloat) * toRateFloat

	transaction := models.Transaction{
		TransactionCode:   transactionCode,
		FromCurrency:      req.FromCurrency,
		ToCurrency:        req.ToCurrency,
		Amount:            req.Amount,
		AmountConverted:   convertedAmount,
		ExchangeRate:      toRateFloat / fromRateFloat,
		TransactionTypeID: transactionTypeID,
		CreatedAt:         time.Now(),
		UserID:            userId,
		Status:            StatusSuccessful,
	}

	transCollection := mongoClient.Database("currencyMongoDb").Collection("transactions")
	result, err := transCollection.InsertOne(context.Background(), transaction)
	if err != nil {
		return saveFailedTransaction(mongoClient, req, StatusFailed, userId, "Failed to save transaction")
	}

	transactionID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.JSON(fiber.Map{
		"transactionId":   transactionID,
		"transactionCode": transaction.TransactionCode,
		"fromCurrency":    req.FromCurrency,
		"toCurrency":      req.ToCurrency,
		"amount":          req.Amount,
		"amountConverted": convertedAmount,
		"exchangeRate":    toRateFloat / fromRateFloat,
		"transactionType": transactionType.Name,
		"createdAt":       transaction.CreatedAt.Format(time.RFC3339),
		"userId":          transaction.UserID,
	})
}

func saveFailedTransaction(mongoClient *mongo.Client, req ConversionRequest, status models.TransactionStatus, userId string, errorMsg string) error {

	transaction := models.Transaction{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Amount:       req.Amount,
		CreatedAt:    time.Now(),
		UserID:       userId,
		Status:       status,
	}

	transCollection := mongoClient.Database("currencyMongoDb").Collection("transactions")
	_, err := transCollection.InsertOne(context.Background(), transaction)
	if err != nil {
		fmt.Println("Failed to save failed transaction:", err)
	}
	return fiber.NewError(fiber.StatusInternalServerError, errorMsg)
}
