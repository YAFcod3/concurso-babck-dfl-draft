// routes/routes.go
package routes

import (
	"exchange-rate/handlers"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(app *fiber.App, mongoClient *mongo.Client) {
	app.Post("/register", func(c *fiber.Ctx) error {
		return handlers.Register(c, mongoClient)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.Login(c, mongoClient)
	})

	// Otras rutas...
}
