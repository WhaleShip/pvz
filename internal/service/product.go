package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/metrics"
)

type productRepoWriter interface {
	InsertProduct(ctx context.Context,
		pvzID, productID uuid.UUID,
		dateTime time.Time,
		productType string) (uuid.UUID, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type productService struct {
	productRepo productRepoWriter
	metrics     metrics.MetricsSender
}

func NewProductService(repo productRepoWriter, aggregator metrics.MetricsSender) *productService {
	return &productService{
		productRepo: repo,
		metrics:     aggregator,
	}
}

func (s *productService) AddProduct(ctx context.Context, req oapi.PostProductsJSONRequestBody) (oapi.Product, error) {
	newProductID := uuid.New()
	now := time.Now()
	receptionID, err := s.productRepo.InsertProduct(ctx, req.PvzId, newProductID, now, string(req.Type))
	if err != nil {
		return oapi.Product{}, err
	}
	product := oapi.Product{
		Id:          &newProductID,
		DateTime:    &now,
		ReceptionId: receptionID,
		Type:        oapi.ProductType(req.Type),
	}

	if s.metrics != nil {
		s.metrics.SendBusinessMetricsUpdate(metrics.MetricsUpdate{
			ProductsAddedDelta: 1,
		})
	}

	return product, nil
}

func (s *productService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return s.productRepo.DeleteLastProduct(ctx, pvzID)
}
