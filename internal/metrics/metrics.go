package metrics

type EndpointMetrics struct {
	HTTPRequestsTotal      int64
	TotalResponseTime      float64
	PvzCreatedTotal        int64
	ReceptionsCreatedTotal int64
	ProductsAddedTotal     int64
}

type MetricsUpdate struct {
	Endpoint               string  `json:"endpoint"`
	HTTPRequestsDelta      int64   `json:"http_requests_delta"`
	ResponseTimeDelta      float64 `json:"response_time_delta"`
	PvzCreatedDelta        int64   `json:"pvz_created_delta"`
	ReceptionsCreatedDelta int64   `json:"receptions_created_delta"`
	ProductsAddedDelta     int64   `json:"products_added_delta"`
}

type MetricsSender interface {
	SendBusinessMetricsUpdate(update MetricsUpdate)
	SendTechMetricsUpdate(update MetricsUpdate)
}
