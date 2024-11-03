package main

import (
	"exchange-rate/database"
	"exchange-rate/handlers"
	"exchange-rate/utils/data_updater"
	"exchange-rate/utils/generate_transaction_code" // Asegúrate de importar el paquete
	"time"

	"github.com/gofiber/fiber/v2"
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

	app.Post("/convert", func(c *fiber.Ctx) error {
		return handlers.ConvertCurrency(c, database.MongoClient, database.RedisClient, codeGen) // Pasar el generador de códigos
	})

	app.Listen(":3000")
}
