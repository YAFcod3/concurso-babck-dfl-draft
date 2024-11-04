package main

import (
	"exchange-rate/database"
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

	// Crear la aplicación Fiber
	app := fiber.New()
	app.Use(cors.New())

	// app.Post("/convert", middleware.RateLimit(5, time.Minute), middleware.IsAuthenticated(), handlers.ConvertCurrency)

	// app.Get("/metrics", func(c *fiber.Ctx) error {
	// 	// Obtener las métricas de Prometheus
	// 	metrics, err := prometheus.DefaultGatherer.Gather()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	metricFamilies := ""
	// 	for _, mf := range metrics {
	// 		metricFamilies += mf.String() + "\n"
	// 	}

	// 	// Establecer el tipo de contenido y devolver las métricas
	// 	c.Set("Content-Type", "text/plain; version=0.0.4")
	// 	c.SendString(metricFamilies)
	// 	return nil
	// })

	// Configurar las rutas
	transactionTypeRepo := repository.NewTransactionTypeRepository(mongoDatabase)
	routes.SetupAuthRoutes(app, database.MongoClient)
	routes.SetupConversionRoutes(app, database.MongoClient, database.RedisClient, codeGen)
	routes.SetupTransactionTypeRoutes(app, transactionTypeRepo)

	// Escuchar en el puerto 8000
	app.Listen(":8000")
}
