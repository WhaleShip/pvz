package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/whaleship/pvz/internal/metrics"
)

type fakeMetrics struct {
	calls []metrics.MetricsUpdate
}

func (f *fakeMetrics) SendTechMetricsUpdate(m metrics.MetricsUpdate) {
	f.calls = append(f.calls, m)
}

func (f *fakeMetrics) SendBusinessMetricsUpdate(m metrics.MetricsUpdate) {
}
func TestMetricsMiddleware(t *testing.T) {
	fm := &fakeMetrics{}
	app := fiber.New()
	app.Use(MetricsMiddleware("TestEndpoint", fm))
	app.Get("/", func(c *fiber.Ctx) error {
		time.Sleep(2 * time.Millisecond)
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	require.Len(t, fm.calls, 1)
	update := fm.calls[0]
	require.Equal(t, "TestEndpoint", update.Endpoint)
	require.Equal(t, int64(1), update.HTTPRequestsDelta)
	require.True(t, update.ResponseTimeDelta >= 0)
}
