package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/metrics"
)

func MetricsMiddleware(handlerName string, ipcManager metrics.MetricsSender) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		ipcManager.SendTechMetricsUpdate(metrics.MetricsUpdate{
			Endpoint:          handlerName,
			HTTPRequestsDelta: 1,
			ResponseTimeDelta: duration,
		})

		return err
	}
}
