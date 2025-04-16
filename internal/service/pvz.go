package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/dto"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/repository"
)

type PVZService interface {
	CreatePVZ(ctx context.Context, req gen.PostPvzJSONRequestBody) (gen.PVZ, error)
	GetPVZ(ctx context.Context, params gen.GetPvzParams) (dto.PVZResponse, error)
}

type pvzService struct {
	pvzRepo       repository.PVZRepository
	receptionRepo repository.ReceptionRepository
	productRepo   repository.ProductRepository
}

func NewPVZService(pvzRepository repository.PVZRepository,
	receptionRepository repository.ReceptionRepository,
	productRepository repository.ProductRepository) PVZService {
	return &pvzService{
		pvzRepo:       pvzRepository,
		receptionRepo: receptionRepository,
		productRepo:   productRepository,
	}
}

func (s *pvzService) CreatePVZ(ctx context.Context, req gen.PostPvzJSONRequestBody) (gen.PVZ, error) {
	switch req.City {
	case gen.Казань, gen.Москва, gen.СанктПетербург:
	default:
		return gen.PVZ{}, pvz_errors.ErrInvalidPVZCity
	}

	now := time.Now()
	pvz, err := s.pvzRepo.InsertPVZ(ctx, req.City, now)
	if err != nil {
		return gen.PVZ{}, fmt.Errorf("%w: %s", pvz_errors.ErrInsertPVZFailed, err.Error())
	}

	return pvz, nil
}

func (s *pvzService) aggregatePVZData(ctx context.Context, pvzList []gen.PVZ) ([]dto.PVZWithReceptions, error) {
	aggregated := make([]dto.PVZWithReceptions, 0, len(pvzList))
	for _, pvz := range pvzList {
		receptions, err := s.receptionRepo.GetReceptionsByPVZ(ctx, *pvz.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", pvz_errors.ErrSelectReceptionsFailed, err.Error())
		}

		var receptionIDs []*uuid.UUID
		for _, rec := range receptions {
			if rec.Id != nil {
				receptionIDs = append(receptionIDs, rec.Id)
			}
		}

		var products []gen.Product
		if len(receptionIDs) > 0 {
			products, err = s.productRepo.GetProductsByReceptionIDs(ctx, receptionIDs)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", pvz_errors.ErrSelectProductsFailed, err.Error())
			}
		}

		productsByReception := make(map[string][]gen.Product)
		for _, prod := range products {
			key := prod.ReceptionId.String()
			productsByReception[key] = append(productsByReception[key], prod)
		}

		var recWithProducts []dto.ReceptionWithProducts
		for _, rec := range receptions {
			key := ""
			if rec.Id != nil {
				key = rec.Id.String()
			}
			recWithProducts = append(recWithProducts, dto.ReceptionWithProducts{
				Reception: rec,
				Products:  productsByReception[key],
			})
		}

		aggregated = append(aggregated, dto.PVZWithReceptions{
			Pvz:        pvz,
			Receptions: recWithProducts,
		})
	}
	return aggregated, nil
}

func (s *pvzService) GetPVZ(ctx context.Context, params gen.GetPvzParams) (dto.PVZResponse, error) {
	pvzList, err := s.pvzRepo.SelectPVZ(ctx, params)
	if err != nil {
		return dto.PVZResponse{}, fmt.Errorf("%w: %s", pvz_errors.ErrSelectPVZFailed, err.Error())
	}

	aggregated, err := s.aggregatePVZData(ctx, pvzList)
	if err != nil {
		return dto.PVZResponse{}, err
	}

	response := dto.PVZResponse{
		PvzWithReceptions: aggregated,
	}
	return response, nil
}
