package service

import "github.com/whaleship/pvz/internal/repository"

type PVZService interface {
}

type pvzService struct {
	pvzRepo repository.PVZRepository
}

func NewPVZService(pvzRepository repository.PVZRepository) PVZService {
	return &pvzService{
		pvzRepo: pvzRepository,
	}
}
