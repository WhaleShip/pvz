package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/metrics"
)

type receptionRepoWriter interface {
	CreateReception(ctx context.Context, req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error)
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (oapi.Reception, error)
}

type receptionService struct {
	receptionRepo receptionRepoWriter
	metrics       metrics.MetricsSender
}

func NewReceptionService(repo receptionRepoWriter, aggregator metrics.MetricsSender) *receptionService {
	return &receptionService{
		receptionRepo: repo,
		metrics:       aggregator,
	}
}
func (s *receptionService) CreateReception(req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error) {
	ctx := context.Background()
	reception, err := s.receptionRepo.CreateReception(ctx, req)
	if err != nil {
		return oapi.Reception{}, err
	}

	s.metrics.SendBusinessMetricsUpdate(metrics.MetricsUpdate{
		ReceptionsCreatedDelta: 1,
	})

	return reception, nil
}

func (s *receptionService) CloseLastReception(pvzID uuid.UUID) (oapi.Reception, error) {
	ctx := context.Background()
	return s.receptionRepo.CloseLastReception(ctx, pvzID)
}
