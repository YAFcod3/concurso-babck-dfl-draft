package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// Init initializes Prometheus metrics
func Init(app *fiber.App) {
	//  agregar métricas personalizadas aquí

	// Configurar la ruta de métricas para Prometheus
	app.Get("/metrics", func(c *fiber.Ctx) error {
		// Obtener las métricas de Prometheus
		metrics, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			return err
		}

		metricFamilies := ""
		for _, mf := range metrics {
			metricFamilies += mf.String() + "\n"
		}

		// Establecer el tipo de contenido y devolver las métricas
		c.Set("Content-Type", "text/plain; version=0.0.4")
		c.SendString(metricFamilies)
		return nil
	})
}
