// handlers/currency.go
package handlers

import (
	"context"
	"exchange-rate/utils/generate_transaction_code"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Transaction struct {
	// TransactionID   string    `bson:"transaction_id"`
	TransactionCode string    `bson:"transaction_code"`
	FromCurrency    string    `bson:"from_currency"`
	ToCurrency      string    `bson:"to_currency"`
	Amount          float64   `bson:"amount"`
	AmountConverted float64   `bson:"amount_converted"`
	ExchangeRate    float64   `bson:"exchange_rate"`
	TransactionType string    `bson:"transaction_type"`
	CreatedAt       time.Time `bson:"created_at"`
	UserID          string    `bson:"user_id"`
}

// ConvertCurrency maneja la conversión de moneda.
func ConvertCurrency(c *fiber.Ctx, mongoClient *mongo.Client, redisClient *redis.Client, codeGen *generate_transaction_code.CodeGenerator) error {
	type ConversionRequest struct {
		FromCurrency    string  `json:"fromCurrency"`
		ToCurrency      string  `json:"toCurrency"`
		Amount          float64 `json:"amount"`
		TransactionType string  `json:"transactionType"`
	}

	var req ConversionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request",
		})
	}

	// Validar campos vacíos
	if req.FromCurrency == "" || req.ToCurrency == "" || req.TransactionType == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "All fields must be filled in and amount must be greater than zero",
		})
	}

	// Verificar si el `transactionType` existe en MongoDB
	// collection := mongoClient.Database("currencyMongoDb").Collection("transaction_types")
	// filter := bson.M{"_id": req.TransactionType}
	// var result bson.M
	// err := collection.FindOne(context.Background(), filter).Decode(&result)
	// if err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"code":    "INVALID_TRANSACTION_TYPE",
	// 		"message": "The transaction type ID is invalid. Please provide a valid transaction type.",
	// 	})
	// }

	// Obtener

	// Obtener la tasa de cambio desde Redis para la moneda de origen
	fromRateFloat := 1.0 // Valor por defecto si la moneda de origen es USD
	if req.FromCurrency != "USD" {
		fromRate, err := redisClient.HGet(context.Background(), "exchange_rates", req.FromCurrency).Result()
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "From currency not found"})
		}
		fromRateFloat, err = strconv.ParseFloat(fromRate, 64)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid from currency rate"})
		}
	}

	// Obtener la tasa de cambio desde Redis para la moneda de destino
	toRateFloat := 1.0 // Valor por defecto si la moneda de destino es USD
	if req.ToCurrency != "USD" {
		toRate, err := redisClient.HGet(context.Background(), "exchange_rates", req.ToCurrency).Result()
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "To currency not found"})
		}
		toRateFloat, err = strconv.ParseFloat(toRate, 64)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid to currency rate"})
		}
	}

	userId := c.Locals("userId").(string)
	fmt.Println("userId : ", userId)

	// Generar el código de transacción
	transactionCode, err := codeGen.GenerateCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "TRANSACTION_CODE_ERROR",
			"message": err.Error(),
		})
	}
	convertedAmount := (req.Amount / fromRateFloat) * toRateFloat
	fmt.Println("transactionCode : ", transactionCode)
	fmt.Println("convertedAmount : ", convertedAmount)

	// Obtener el ID generado por MongoDB

	transaction := Transaction{
		// TransactionID:   generate_transaction_code.GenerateUniqueID(),
		TransactionCode: transactionCode,
		FromCurrency:    req.FromCurrency,
		ToCurrency:      req.ToCurrency,
		Amount:          req.Amount,
		AmountConverted: convertedAmount,
		ExchangeRate:    toRateFloat / fromRateFloat,
		TransactionType: req.TransactionType,
		CreatedAt:       time.Now(),
		UserID:          userId,
		// UserID:          "id del usuario que hizo la transacción", // Obtén este valor de tu lógica de usuario
	}

	// Guardar la transacción en MongoDB
	transCollection := mongoClient.Database("currencyMongoDb").Collection("transactions")
	result, err := transCollection.InsertOne(context.Background(), transaction)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save transaction"})
	}

	transactionID := result.InsertedID.(primitive.ObjectID).Hex()

	return c.JSON(fiber.Map{
		// "transactionId":   transactionID,
		"transactionId":   transactionID,
		"transactionCode": transaction.TransactionCode,
		"fromCurrency":    req.FromCurrency,
		"toCurrency":      req.ToCurrency,
		"amount":          req.Amount,
		"amountConverted": convertedAmount,
		"exchangeRate":    toRateFloat / fromRateFloat,
		"createdAt":       transaction.CreatedAt.Format(time.RFC3339),
		"userId":          transaction.UserID,
	})
}
