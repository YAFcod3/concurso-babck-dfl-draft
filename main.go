package main

import (
	"exchange-rate/database"
	"exchange-rate/handlers"
	"exchange-rate/middleware"
	"exchange-rate/repository"
	"exchange-rate/utils/data_updater"
	"exchange-rate/utils/generate_transaction_code" // Asegúrate de importar el paquete
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.Init()
	defer database.Close()

	// Iniciar actualizador de tasas de cambio
	data_updater.StartExchangeRateUpdater(database.RedisClient, 1*time.Hour)

	// Crear una instancia de CodeGenerator
	codeGen := &generate_transaction_code.CodeGenerator{Client: database.RedisClient}
	codeGen.LoadLastCounter()

	app := fiber.New()
	app.Use(cors.New())

	// app.Post("/convert", middleware.RateLimit(5, time.Minute), middleware.IsAuthenticated(), handlers.ConvertCurrency)

	// app.Use(middleware.RateLimit(5, time.Minut))
	// routes.SetupRoutes(app, mongoClient, database.RedisClient, codeGen)
	mongoDatabase := database.MongoClient.Database("currencyMongoDb")

	transactionTypeRepo := repository.NewTransactionTypeRepository(mongoDatabase)

	app.Post("/register", func(c *fiber.Ctx) error {
		return handlers.Register(c, database.MongoClient)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.Login(c, database.MongoClient)
	})
	// app.Post("/convert", middleware.IsAuthenticated(), middleware.VerifyTransactionDuplicated(database.RedisClient), func(c *fiber.Ctx) error {
	// 	return handlers.ConvertCurrency(c, database.MongoClient, database.RedisClient, codeGen)
	// })
	app.Post("/convert", middleware.IsAuthenticated(), func(c *fiber.Ctx) error {
		return handlers.ConvertCurrency(c, database.MongoClient, database.RedisClient, codeGen)
	})

	app.Post("/api/settings/transactions-types", func(c *fiber.Ctx) error {
		return handlers.CreateTransactionType(c, transactionTypeRepo)
	})
	app.Get("/api/settings/transactions-types", func(c *fiber.Ctx) error {
		return handlers.GetTransactionTypes(c, transactionTypeRepo)
	})
	app.Put("/api/settings/transactions-types/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateTransactionType(c, transactionTypeRepo)
	})
	app.Delete("/api/settings/transactions-types/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteTransactionType(c, transactionTypeRepo)
	})

	app.Listen(":3000")
}
