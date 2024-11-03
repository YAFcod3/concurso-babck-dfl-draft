package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

var requests = make(map[string]int)

func RateLimit(maxRequests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if requests[ip] > maxRequests {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many requests"})
		}
		requests[ip]++
		time.AfterFunc(window, func() {
			requests[ip]--
		})
		return c.Next()
	}
}
