package service

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/whaleship/pvz/internal/metrics"
)

func uuidPtr(u uuid.UUID) *uuid.UUID { return &u }

type mockMetrics struct{ mock.Mock }

func (m *mockMetrics) SendBusinessMetricsUpdate(u metrics.MetricsUpdate) {
	m.Called(u)
}

func (m *mockMetrics) SendTechMetricsUpdate(u metrics.MetricsUpdate) {
	m.Called(u)
}
