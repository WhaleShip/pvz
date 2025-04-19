package grpc_handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/gen/proto"
)

type mockPVZService struct{ mock.Mock }

func (m *mockPVZService) GetAllPVZs(ctx context.Context) ([]*proto.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*proto.PVZ), args.Error(1)
}

func TestGetPVZList(t *testing.T) {
	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		handler := NewPVZGRPCService(mockSvc)
		ctx := context.Background()
		req := &proto.GetPVZListRequest{}

		mockSvc.
			On("GetAllPVZs", ctx).
			Return([]*proto.PVZ(nil), errors.New("failed to fetch"))

		resp, err := handler.GetPVZList(ctx, req)
		require.Nil(t, resp)
		require.EqualError(t, err, "failed to fetch")
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		handler := NewPVZGRPCService(mockSvc)
		ctx := context.Background()
		req := &proto.GetPVZListRequest{}

		expected := []*proto.PVZ{
			{Id: "pvz1", City: "Moscow"},
			{Id: "pvz2", City: "Kazan"},
		}
		mockSvc.
			On("GetAllPVZs", ctx).
			Return(expected, nil)

		resp, err := handler.GetPVZList(ctx, req)
		require.NoError(t, err)
		require.Equal(t, expected, resp.Pvzs)
		mockSvc.AssertExpectations(t)
	})
}
