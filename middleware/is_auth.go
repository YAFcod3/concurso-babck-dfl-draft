package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func IsAuthenticated() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Parsear el token
		claims := &jwt.StandardClaims{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("your_secret_key"), nil // Cambia "your_secret_key" por tu clave secreta
		})

		if err != nil || !tkn.Valid {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// Almacenar el userId en los datos locales para el contexto
		c.Locals("userId", claims.Subject)
		return c.Next()
	}
}
