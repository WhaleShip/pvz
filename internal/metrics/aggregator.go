package metrics

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Aggregator struct {
	mu      sync.Mutex
	metrics map[string]*EndpointMetrics
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		metrics: make(map[string]*EndpointMetrics),
	}
}

func (a *Aggregator) UpdateMetrics(update MetricsUpdate) {
	a.mu.Lock()
	defer a.mu.Unlock()

	m, exists := a.metrics[update.Endpoint]
	if !exists {
		m = &EndpointMetrics{}
		a.metrics[update.Endpoint] = m
	}
	m.HTTPRequestsTotal += update.HTTPRequestsDelta
	m.TotalResponseTime += update.ResponseTimeDelta
	m.PvzCreatedTotal += update.PvzCreatedDelta
	m.ReceptionsCreatedTotal += update.ReceptionsCreatedDelta
	m.ProductsAddedTotal += update.ProductsAddedDelta
}

func (a *Aggregator) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		output := ""
		for endpoint, m := range a.metrics {
			if endpoint != "" {
				output += fmt.Sprintf("http_requests_total{endpoint=\"%s\"} %d\n", endpoint, m.HTTPRequestsTotal)
				output += fmt.Sprintf("http_request_duration_total{endpoint=\"%s\"} %f\n", endpoint, m.TotalResponseTime)
			}
		}
		if mGlobal, exists := a.metrics[""]; exists {
			output += fmt.Sprintf("pvz_created_total %d\n", mGlobal.PvzCreatedTotal)
			output += fmt.Sprintf("receptions_created_total %d\n", mGlobal.ReceptionsCreatedTotal)
			output += fmt.Sprintf("products_added_total %d\n", mGlobal.ProductsAddedTotal)
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if _, err := w.Write([]byte(output)); err != nil {
			log.Println(err)
		}
	}
}
