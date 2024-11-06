package middleware

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duración de las solicitudes HTTP en segundos.",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10},
		},
		[]string{"method", "path"},
	)

	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Número total de solicitudes HTTP.",
		},
		[]string{"method", "path", "status"},
	)

	errorRequestsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_requests_errors_total",
			Help: "Total number of error requests processed by the MyApp web server",
		},
		[]string{"method", "path", "status"},
	)

	//  métricas específicas para Redis y MongoDB
	// databaseDuration = prometheus.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Name:    "database_operation_duration_seconds",
	// 		Help:    "Duración de las operaciones de base de datos en segundos.",
	// 		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5},
	// 	},
	// 	[]string{"database", "operation"},
	// )

	// cacheDuration = prometheus.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Name:    "cache_operation_duration_seconds",
	// 		Help:    "Duración de las operaciones de caché en segundos.",
	// 		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
	// 	},
	// 	[]string{"operation"},
	// )
)

func RegisterPrometheus(app *fiber.App) {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(errorRequestsCount)

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		path := c.Route().Path

		requestDuration.WithLabelValues(c.Method(), path).Observe(duration)
		totalRequests.WithLabelValues(
			c.Method(),
			path,
			strconv.Itoa(c.Response().StatusCode()),
		).Inc()
		errorRequestsCount.WithLabelValues(
			c.Method(),
			path,
			strconv.Itoa(c.Response().StatusCode()),
		).Inc()

		return err
	})

	// app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/metrics", func(c *fiber.Ctx) error {
		log.Printf("Metrics endpoint accessed from: %s", c.IP())
		return adaptor.HTTPHandler(promhttp.Handler())(c)
	})

	// Añadir un endpoint de health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
}

// Funciones auxiliares para registrar operaciones de DB y caché
// func ObserveDatabaseOperation(database, operation string, duration time.Duration) {
// 	databaseDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
// }

// func ObserveCacheOperation(operation string, duration time.Duration) {
// 	cacheDuration.WithLabelValues(operation).Observe(duration.Seconds())
// }
