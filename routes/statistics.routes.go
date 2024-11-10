package routes

import (
	"exchange-rate/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupStatisticsRoutes(app *fiber.App, mongoClient *mongo.Client) {
	statisticsGroup := app.Group("/api/statistics/")

	statisticsGroup.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetStatistics(c, mongoClient)
	})

}
