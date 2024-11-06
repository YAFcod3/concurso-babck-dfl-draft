package main

import (
	"exchange-rate/database"
	"exchange-rate/handlers"
	"exchange-rate/middleware"
	"exchange-rate/repository"
	"exchange-rate/routes"
	"exchange-rate/utils/data_updater"
	"exchange-rate/utils/generate_transaction_code"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"time"
)

func main() {
	database.Init()
	defer database.Close()
	mongoDatabase := database.MongoClient.Database("currencyMongoDb")

	data_updater.StartExchangeRateUpdater(database.RedisClient, 1*time.Hour)

	codeGen := &generate_transaction_code.CodeGenerator{Client: database.RedisClient}
	codeGen.LoadLastCounter()

	app := fiber.New()
	app.Use(cors.New())

	middleware.RegisterPrometheus(app)

	transactionTypeRepo := repository.NewTransactionTypeRepository(mongoDatabase)
	routes.SetupAuthRoutes(app, database.MongoClient)
	routes.SetupConversionRoutes(app, database.MongoClient, database.RedisClient, codeGen)
	routes.SetupTransactionTypeRoutes(app, transactionTypeRepo)

	app.Get("/api/statistics", func(c *fiber.Ctx) error {
		return handlers.GetStatistics(c, database.MongoClient)
	})

	app.Listen(":8000")
}
