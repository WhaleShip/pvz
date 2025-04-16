package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/repository"
)

type ReceptionService interface {
	CreateReception(req gen.PostReceptionsJSONRequestBody) (gen.Reception, error)
	CloseLastReception(pvzID uuid.UUID) (gen.Reception, error)
}

type receptionService struct {
	receptionRepo repository.ReceptionRepository
}

func NewReceptionService(repo repository.ReceptionRepository) ReceptionService {
	return &receptionService{
		receptionRepo: repo,
	}
}
func (s *receptionService) CreateReception(req gen.PostReceptionsJSONRequestBody) (gen.Reception, error) {
	ctx := context.Background()
	reception, err := s.receptionRepo.CreateReception(ctx, req)
	if err != nil {
		return gen.Reception{}, err
	}

	return reception, nil
}

func (s *receptionService) CloseLastReception(pvzID uuid.UUID) (gen.Reception, error) {
	ctx := context.Background()
	return s.receptionRepo.CloseLastReception(ctx, pvzID)
}
