package routes

import (
	"exchange-rate/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupAuthRoutes(app *fiber.App, mongoClient *mongo.Client) {
	authGroup := app.Group("/api/auth")

	authGroup.Post("/register", func(c *fiber.Ctx) error {
		return handlers.Register(c, mongoClient)
	})

	authGroup.Post("/login", func(c *fiber.Ctx) error {
		return handlers.Login(c, mongoClient)
	})
}
