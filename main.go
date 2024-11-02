package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"exchange-rate/utils/data_updater"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Conexión a Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Conexión a MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Iniciar actualizador de tasas de cambio
	data_updater.StartExchangeRateUpdater(client, 1*time.Hour)

	app := fiber.New()

	app.Post("/convert", func(c *fiber.Ctx) error {
		type ConversionRequest struct {
			FromCurrency    string  `json:"fromCurrency"`
			ToCurrency      string  `json:"toCurrency"`
			Amount          float64 `json:"amount"`
			TransactionType string  `json:"transactionType"`
		}

		var req ConversionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Verificar si el `transactionType` existe en MongoDB
		// collection := mongoClient.Database("your_database").Collection("transaction_types")
		// filter := bson.M{"_id": req.TransactionType}
		// var result bson.M
		// err := collection.FindOne(context.Background(), filter).Decode(&result)
		// if err != nil {
		// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		// 		"code":    "INVALID_TRANSACTION_TYPE",
		// 		"message": "The transaction type ID is invalid. Please provide a valid transaction type.",
		// 	})
		// }

		// Obtener la tasa de cambio desde Redis para la moneda de origen
		fromRateFloat := 1.0 // Valor por defecto si la moneda de origen es USD
		if req.FromCurrency != "USD" {
			fromRate, err := client.HGet(context.Background(), "exchange_rates", req.FromCurrency).Result()
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "From currency not found"})
			}
			fmt.Println("fromRate" + fromRate)
			fromRateFloat, err = strconv.ParseFloat(fromRate, 64)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid from currency rate"})
			}
		}

		// Obtener la tasa de cambio desde Redis para la moneda de destino
		toRateFloat := 1.0 // Valor por defecto si la moneda de destino es USD
		if req.ToCurrency != "USD" {
			toRate, err := client.HGet(context.Background(), "exchange_rates", req.ToCurrency).Result()
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "To currency not found"})
			}
			fmt.Println("toRate" + toRate)

			toRateFloat, err = strconv.ParseFloat(toRate, 64)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid to currency rate"})
			}
			fmt.Println("toRateFloat", toRateFloat)

		}

		convertedAmount := (req.Amount / fromRateFloat) * toRateFloat

		// Generar el código de transacción (aquí puedes añadir tu lógica de generación de código)

		//todo ok entonces guardar en base de datos estos dettales para desp tener un historial de transacciones y oitras estadisticas necesarias

		return c.JSON(fiber.Map{
			"transactionId":   "abc123def456",
			"transactionCode": "T24101811210001",
			"fromCurrency":    req.FromCurrency,
			"toCurrency":      req.ToCurrency,
			"amount":          req.Amount,
			"amountConverted": convertedAmount,
			"exchangeRate":    toRateFloat / fromRateFloat,
			// "transactionType": result["name"], // supón que 'name' es el nombre del tipo de transacción en MongoDB
			"createdAt": time.Now().Format(time.RFC3339),
			"userId":    "id del usuario que hizo la transacción", // lo obtengo por el middleware y el jwt
		})
	})

	app.Listen(":3000")
}
