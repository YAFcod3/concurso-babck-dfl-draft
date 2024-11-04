package main

import (
	"exchange-rate/database"
	"exchange-rate/handlers"
	"exchange-rate/repository"
	"exchange-rate/routes"
	"exchange-rate/utils/data_updater"
	"exchange-rate/utils/generate_transaction_code"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Inicializar la base de datos
	database.Init()
	defer database.Close()
	mongoDatabase := database.MongoClient.Database("currencyMongoDb")

	// Iniciar actualizador de tasas de cambio
	data_updater.StartExchangeRateUpdater(database.RedisClient, 1*time.Hour)

	// Crear una instancia de CodeGenerator
	codeGen := &generate_transaction_code.CodeGenerator{Client: database.RedisClient}
	codeGen.LoadLastCounter()

	app := fiber.New()
	app.Use(cors.New())

	// prometheus
	// metrics.Init(app)
	// app.Post("/convert", middleware.RateLimit(5, time.Minute), middleware.IsAuthenticated(), handlers.ConvertCurrency)
	// Configurar las rutas
	transactionTypeRepo := repository.NewTransactionTypeRepository(mongoDatabase)
	routes.SetupAuthRoutes(app, database.MongoClient)
	routes.SetupConversionRoutes(app, database.MongoClient, database.RedisClient, codeGen)
	routes.SetupTransactionTypeRoutes(app, transactionTypeRepo)
	// ! esta mal statistics
	app.Get("/api/statistics", func(c *fiber.Ctx) error {
		return handlers.GetStatistics(c, database.MongoClient)
	})
	//      Path:/api/transactions     historial d trnasacciones
	//   Método:GET

	//      Path: /api/currencies
	//   Método: GET

	app.Listen(":8000")
}
