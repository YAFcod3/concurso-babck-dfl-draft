package routes

import (
	"exchange-rate/handlers"
	"exchange-rate/repository"

	"github.com/gofiber/fiber/v2"
)

func SetupTransactionTypeRoutes(app *fiber.App, transactionTypeRepo *repository.TransactionTypeRepository) {
	transactionTypeGroup := app.Group("/api/settings/transactions-types")

	transactionTypeGroup.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetTransactionTypes(c, transactionTypeRepo)
	})

	transactionTypeGroup.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateTransactionType(c, transactionTypeRepo)
	})

	transactionTypeGroup.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateTransactionType(c, transactionTypeRepo)
	})

	transactionTypeGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteTransactionType(c, transactionTypeRepo)
	})
}
