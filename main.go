package main

import (
	"exchange-rate/database"
	"exchange-rate/middleware"
	"exchange-rate/repository"
	"exchange-rate/routes"
	"exchange-rate/utils/data_updater"
	"exchange-rate/utils/generate_transaction_code"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	routes.RegisterRoutes(app, database.MongoClient, database.RedisClient, codeGen, transactionTypeRepo)

	app.Listen(":8000")
}
