package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateMetrics(t *testing.T) {
	t.Run("single update", func(t *testing.T) {
		a := NewAggregator()
		update := MetricsUpdate{
			Endpoint:               "/foo",
			HTTPRequestsDelta:      5,
			ResponseTimeDelta:      1.23,
			PvzCreatedDelta:        2,
			ReceptionsCreatedDelta: 3,
			ProductsAddedDelta:     4,
		}
		a.UpdateMetrics(update)
		m := a.metrics["/foo"]
		require.Equal(t, int64(5), m.HTTPRequestsTotal)
		require.InEpsilon(t, 1.23, m.TotalResponseTime, 1e-6)
		require.Equal(t, int64(2), m.PvzCreatedTotal)
		require.Equal(t, int64(3), m.ReceptionsCreatedTotal)
		require.Equal(t, int64(4), m.ProductsAddedTotal)
	})

	t.Run("multiple updates accumulate", func(t *testing.T) {
		a := NewAggregator()
		a.UpdateMetrics(MetricsUpdate{Endpoint: "/bar", HTTPRequestsDelta: 1, ResponseTimeDelta: 0.5})
		a.UpdateMetrics(MetricsUpdate{Endpoint: "/bar", HTTPRequestsDelta: 2, ResponseTimeDelta: 0.25})
		m := a.metrics["/bar"]
		require.Equal(t, int64(3), m.HTTPRequestsTotal)
		require.InEpsilon(t, 0.75, m.TotalResponseTime, 1e-6)
	})
}

func TestHTTPHandler(t *testing.T) {
	t.Run("renders per-endpoint and global metrics", func(t *testing.T) {
		a := NewAggregator()
		a.UpdateMetrics(MetricsUpdate{Endpoint: "/x", HTTPRequestsDelta: 7, ResponseTimeDelta: 0.7})
		a.UpdateMetrics(MetricsUpdate{Endpoint: "", PvzCreatedDelta: 1, ReceptionsCreatedDelta: 2, ProductsAddedDelta: 3})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler := a.HTTPHandler()
		handler(w, req)
		body := w.Body.String()

		require.Equal(t, "text/plain; version=0.0.4", w.Header().Get("Content-Type"))
		require.True(t, strings.Contains(body, `http_requests_total{endpoint="/x"} 7`))
		require.True(t, strings.Contains(body, `http_request_duration_total{endpoint="/x"} 0.700000`))
		require.True(t, strings.Contains(body, `pvz_created_total 1`))
		require.True(t, strings.Contains(body, `receptions_created_total 2`))
		require.True(t, strings.Contains(body, `products_added_total 3`))
	})
}
