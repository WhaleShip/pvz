package service

import (
	"context"
	"fmt"
	"time"

	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/repository"
)

type PVZService interface {
	CreatePVZ(req gen.PostPvzJSONRequestBody) (gen.PVZ, error)
}

type pvzService struct {
	pvzRepo repository.PVZRepository
}

func NewPVZService(pvzRepository repository.PVZRepository) PVZService {
	return &pvzService{
		pvzRepo: pvzRepository,
	}
}

func (s *pvzService) CreatePVZ(req gen.PostPvzJSONRequestBody) (gen.PVZ, error) {
	switch req.City {
	case gen.Казань, gen.Москва, gen.СанктПетербург:
	default:
		return gen.PVZ{}, pvz_errors.ErrInvalidPVZCity
	}

	now := time.Now()
	pvz, err := s.pvzRepo.InsertPVZ(context.Background(), req.City, now)
	if err != nil {
		return gen.PVZ{}, fmt.Errorf("%w: %s", pvz_errors.ErrInsertPVZFailed, err.Error())
	}

	return pvz, nil
}
