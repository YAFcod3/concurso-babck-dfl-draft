package routes

import (
	"exchange-rate/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupTransactionsHistoryRoutes(app *fiber.App, mongoClient *mongo.Client) {
	statisticsGroup := app.Group("/api/transactions")

	statisticsGroup.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetTransactions(c, mongoClient)
	})

}
