package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/dto"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/gen/proto"
	"github.com/whaleship/pvz/internal/metrics"
)

type productRepoReader interface {
	GetProductsByReceptionIDs(ctx context.Context, receptionIDs []*uuid.UUID) ([]oapi.Product, error)
}

type receptionRepoReader interface {
	GetReceptionsByPVZ(ctx context.Context, pvzID uuid.UUID) ([]dto.Reception, error)
}

type pvzRepository interface {
	InsertPVZ(ctx context.Context, city oapi.PVZCity, registrationDate time.Time) (oapi.PVZ, error)
	SelectPVZByOpenReceptions(
		ctx context.Context,
		startDate, endDate time.Time,
		limit, offset int,
	) ([]oapi.PVZ, error)
	SelectAllPVZs(ctx context.Context) ([]*proto.PVZ, error)
}

type pvzService struct {
	pvzRepo       pvzRepository
	receptionRepo receptionRepoReader
	productRepo   productRepoReader
	metrics       metrics.MetricsSender
}

func NewPVZService(
	pvzRepository pvzRepository,
	receptionRepository receptionRepoReader,
	productRepository productRepoReader,
	aggregator metrics.MetricsSender,
) *pvzService {
	return &pvzService{
		pvzRepo:       pvzRepository,
		receptionRepo: receptionRepository,
		productRepo:   productRepository,
		metrics:       aggregator,
	}
}

func (s *pvzService) CreatePVZ(ctx context.Context, req oapi.PostPvzJSONRequestBody) (oapi.PVZ, error) {
	switch req.City {
	case oapi.Казань, oapi.Москва, oapi.СанктПетербург:
	default:
		return oapi.PVZ{}, pvz_errors.ErrInvalidPVZCity
	}

	now := time.Now()
	pvz, err := s.pvzRepo.InsertPVZ(ctx, req.City, now)
	if err != nil {
		return oapi.PVZ{}, fmt.Errorf("%w: %s", pvz_errors.ErrInsertPVZFailed, err.Error())
	}

	s.metrics.SendBusinessMetricsUpdate(metrics.MetricsUpdate{
		PvzCreatedDelta: 1,
	})

	return pvz, nil
}

func (s *pvzService) aggregatePVZData(ctx context.Context, pvzList []oapi.PVZ) ([]dto.PVZWithReceptions, error) {
	aggregated := make([]dto.Reception, 0, len(pvzList))
	for _, pvz := range pvzList {
		recs, err := s.receptionRepo.GetReceptionsByPVZ(ctx, *pvz.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", pvz_errors.ErrSelectReceptionsFailed, err)
		}
		aggregated = append(aggregated, recs...)
	}

	var receptions []*uuid.UUID
	for _, r := range aggregated {
		if r.Id != nil {
			receptions = append(receptions, r.Id)
		}
	}

	var products []oapi.Product
	if len(receptions) > 0 {
		var err error
		products, err = s.productRepo.GetProductsByReceptionIDs(ctx, receptions)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", pvz_errors.ErrSelectProductsFailed, err)
		}
	}

	prodByRec := make(map[string][]oapi.Product, len(products))
	for _, p := range products {
		key := p.ReceptionId.String()
		prodByRec[key] = append(prodByRec[key], p)
	}

	result := make([]dto.PVZWithReceptions, 0, len(pvzList))
	for _, pvz := range pvzList {
		var group []dto.ReceptionWithProducts
		for _, r := range aggregated {
			if r.PvzId != *pvz.Id {
				continue
			}
			key := ""
			if r.Id != nil {
				key = r.Id.String()
			}
			group = append(group, dto.ReceptionWithProducts{
				Reception: r,
				Products:  prodByRec[key],
			})
		}
		result = append(result, dto.PVZWithReceptions{
			Pvz:        pvz,
			Receptions: group,
		})
	}

	return result, nil
}

func (s *pvzService) GetPVZ(ctx context.Context, params oapi.GetPvzParams) ([]dto.PVZWithReceptions, error) {
	page, limit := 1, 10
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	offset := (page - 1) * limit
	var startDate, endDate time.Time

	if params.StartDate != nil {
		startDate = *params.StartDate
	} else {
		startDate = time.Time{}
	}
	if params.EndDate != nil {
		endDate = *params.EndDate
	} else {
		endDate = time.Now()
	}
	pvzList, err := s.pvzRepo.SelectPVZByOpenReceptions(
		ctx,
		startDate,
		endDate,
		limit,
		offset,
	)
	if err != nil {
		return []dto.PVZWithReceptions{}, fmt.Errorf("%w: %w", pvz_errors.ErrSelectPVZFailed, err)
	}

	aggregated, err := s.aggregatePVZData(ctx, pvzList)
	if err != nil {
		return []dto.PVZWithReceptions{}, err
	}

	return aggregated, nil
}

func (s *pvzService) GetAllPVZs(ctx context.Context) ([]*proto.PVZ, error) {
	pvzs, err := s.pvzRepo.SelectAllPVZs(ctx)
	if err != nil {
		return nil, err
	}
	return pvzs, nil
}
