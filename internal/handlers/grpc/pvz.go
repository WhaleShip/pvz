package grpc_handlers

import (
	"context"

	"github.com/whaleship/pvz/internal/gen/proto"
)

type PVZService interface {
	GetAllPVZs(ctx context.Context) ([]*proto.PVZ, error)
}

type PVZGRPCService struct {
	proto.UnimplementedPVZServiceServer
	pvzSvc PVZService
}

func NewPVZGRPCService(pvzSvc PVZService) *PVZGRPCService {
	return &PVZGRPCService{
		pvzSvc: pvzSvc,
	}
}

func (s *PVZGRPCService) GetPVZList(
	ctx context.Context,
	req *proto.GetPVZListRequest) (*proto.GetPVZListResponse, error) {
	pvzs, err := s.pvzSvc.GetAllPVZs(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.GetPVZListResponse{
		Pvzs: pvzs,
	}, nil
}
