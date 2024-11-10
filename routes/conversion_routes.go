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

	conversionGroup.Post("/", middleware.IsAuthenticated(),
		// middleware.VerifyTransactionDuplicated(redisClient),  // ! mejorar esto (tiene q coinicir exactamente las 3 Si un usuario realiza una transacción idéntica (mismo monto, tipode transacción y monedas),
		func(c *fiber.Ctx) error {
			return handlers.ConvertCurrency(c, mongoClient, redisClient, codeGen)
		})
}
