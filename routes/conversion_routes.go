package routes

import (
	"exchange-rate/handlers"
	"exchange-rate/middleware"
	"exchange-rate/utils/generate_transaction_code"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupConversionRoutes(app *fiber.App, mongoClient *mongo.Client, redisClient *redis.Client, codeGen *generate_transaction_code.CodeGenerator) {
	conversionGroup := app.Group("/api/conversion")
	// app.Post("/convert", middleware.RateLimit(5, time.Minute), middleware.IsAuthenticated(), handlers.ConvertCurrency)

	conversionGroup.Post("/", middleware.IsAuthenticated(), middleware.VerifyTransactionDuplicated(redisClient), func(c *fiber.Ctx) error {
		return handlers.ConvertCurrency(c, mongoClient, redisClient, codeGen)
	})
}
