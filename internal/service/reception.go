package service

import "github.com/whaleship/pvz/internal/repository"

type ReceptionService interface {
}

type receptionService struct {
	receptionRepo repository.ReceptionRepository
}

func NewReceptionService(repo repository.ReceptionRepository) ReceptionService {
	return &receptionService{
		receptionRepo: repo,
	}
}
