package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
)

func VerifyTransactionDuplicated(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Obtener los datos de la transacción desde el cuerpo de la solicitud
		var transaction struct {
			FromCurrency    string  `json:"from_currency"`
			ToCurrency      string  `json:"to_currency"`
			Amount          float64 `json:"amount"`
			TransactionType string  `json:"transaction_type"`
		}
		if err := c.BodyParser(&transaction); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid transaction data"})
		}

		// Obtener el UserID desde el contexto (extraído del token en el middleware IsAuthenticated)
		userID, ok := c.Locals("userId").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
		}

		// Generar una clave única para la transacción en Redis
		uniqueKey := fmt.Sprintf("%s:%s:%s:%.2f:%s", userID, transaction.FromCurrency, transaction.ToCurrency, transaction.Amount, transaction.TransactionType)

		// Verificar si la clave ya existe en Redis
		ctx := context.Background()
		exists, err := redisClient.Exists(ctx, uniqueKey).Result()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking duplicate transaction"})
		}

		// Si existe, devolver un mensaje de error de transacción duplicada
		if exists > 0 {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"code":    "DUPLICATE_TRANSACTION",
				"message": "A similar transaction was already processed within the last 20 seconds. Please try again later.",
			})
		}

		// Si no existe, guardar la clave en Redis con un TTL de 20 segundos
		err = redisClient.Set(ctx, uniqueKey, "1", 20*time.Second).Err()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error setting transaction limit in Redis"})
		}

		// Continuar al siguiente handler si no es duplicado
		return c.Next()
	}
}
