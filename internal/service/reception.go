package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
	"github.com/whaleship/pvz/internal/repository"
)

type ReceptionService interface {
	CreateReception(req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error)
	CloseLastReception(pvzID uuid.UUID) (oapi.Reception, error)
}

type receptionService struct {
	receptionRepo repository.ReceptionRepository
	metrics       *infrastructure.IPCManager
}

func NewReceptionService(repo repository.ReceptionRepository, aggregator *infrastructure.IPCManager) ReceptionService {
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

	s.metrics.ReportMetrics(metrics.MetricsUpdate{
		ReceptionsCreatedDelta: 1,
	})

	return reception, nil
}

func (s *receptionService) CloseLastReception(pvzID uuid.UUID) (oapi.Reception, error) {
	ctx := context.Background()
	return s.receptionRepo.CloseLastReception(ctx, pvzID)
}
