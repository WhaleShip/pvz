package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
	"github.com/whaleship/pvz/internal/repository"
)

type ProductService interface {
	AddProduct(ctx context.Context, req gen.PostProductsJSONRequestBody) (gen.Product, error)
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

func (s *productService) AddProduct(ctx context.Context, req gen.PostProductsJSONRequestBody) (gen.Product, error) {
	newProductID := uuid.New()
	now := time.Now()
	receptionID, err := s.productRepo.InsertProduct(ctx, req.PvzId, newProductID, now, string(req.Type))
	if err != nil {
		return gen.Product{}, err
	}
	product := gen.Product{
		Id:          &newProductID,
		DateTime:    &now,
		ReceptionId: receptionID,
		Type:        gen.ProductType(req.Type),
	}

	s.metrics.ReportMetrics(metrics.MetricsUpdate{
		ProductsAddedDelta: 1,
	})

	return product, nil
}

func (s *productService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return s.productRepo.DeleteLastProduct(ctx, pvzID)
}
