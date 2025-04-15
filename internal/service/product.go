package service

import "github.com/whaleship/pvz/internal/repository"

type ProductService interface {
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: repo,
	}
}
