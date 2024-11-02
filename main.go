package main

import (
	"fmt"
	"time"

	"exchange-rate/utils/data_updater"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Inicializa el cliente de Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	data_updater.StartExchangeRateUpdater(client, 1*time.Minute)

	app := fiber.New()

	app.Get("/exchange-rates/:currency", func(c *fiber.Ctx) error {
		currency := c.Params("currency")
		rate, err := client.Get(c.Context(), currency).Float64()
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Exchange rate for %s not found", currency),
			})
		}
		return c.JSON(fiber.Map{
			"currency": currency,
			"rate":     rate,
		})
	})

	app.Listen(":3000")
}
