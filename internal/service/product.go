package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
	"github.com/whaleship/pvz/internal/repository"
)

type ProductService interface {
	AddProduct(ctx context.Context, req oapi.PostProductsJSONRequestBody) (oapi.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type productService struct {
	productRepo repository.ProductRepository
	metrics     *infrastructure.IPCManager
}

func NewProductService(repo repository.ProductRepository, aggregator *infrastructure.IPCManager) ProductService {
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

	s.metrics.ReportMetrics(metrics.MetricsUpdate{
		ProductsAddedDelta: 1,
	})

	return product, nil
}

func (s *productService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return s.productRepo.DeleteLastProduct(ctx, pvzID)
}
