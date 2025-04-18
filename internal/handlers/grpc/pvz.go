package grpc_handlers

import (
	"context"

	"github.com/whaleship/pvz/internal/gen/proto"
	"github.com/whaleship/pvz/internal/service"
)

type PVZGRPCService struct {
	proto.UnimplementedPVZServiceServer
	pvzSvc service.PVZService
}

func NewPVZGRPCService(pvzSvc service.PVZService) *PVZGRPCService {
	return &PVZGRPCService{
		pvzSvc: pvzSvc,
	}
}

func (s *PVZGRPCService) GetPVZList(ctx context.Context, req *proto.GetPVZListRequest) (*proto.GetPVZListResponse, error) {
	pvzs, err := s.pvzSvc.GetAllPVZs(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.GetPVZListResponse{
		Pvzs: pvzs,
	}, nil
}
