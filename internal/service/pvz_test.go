package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/dto"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/gen/proto"
	"github.com/whaleship/pvz/internal/metrics"
)

type mockPVZRepo struct{ mock.Mock }

func (m *mockPVZRepo) InsertPVZ(ctx context.Context, city oapi.PVZCity, registrationDate time.Time) (oapi.PVZ, error) {
	args := m.Called(ctx, city, registrationDate)
	return args.Get(0).(oapi.PVZ), args.Error(1)
}

func (m *mockPVZRepo) SelectPVZByOpenReceptions(
	ctx context.Context,
	startDate, endDate time.Time,
	limit, offset int) ([]oapi.PVZ, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]oapi.PVZ), args.Error(1)
}

func (m *mockPVZRepo) SelectAllPVZs(ctx context.Context) ([]*proto.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*proto.PVZ), args.Error(1)
}

type mockReceptionReader struct{ mock.Mock }

func (m *mockReceptionReader) GetReceptionsByPVZ(ctx context.Context, pvzID uuid.UUID) ([]dto.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).([]dto.Reception), args.Error(1)
}

type mockProductReader struct{ mock.Mock }

func (m *mockProductReader) GetProductsByReceptionIDs(
	ctx context.Context,
	receptionIDs []*uuid.UUID) ([]oapi.Product, error) {
	args := m.Called(ctx, receptionIDs)
	return args.Get(0).([]oapi.Product), args.Error(1)
}

func TestCreatePVZ(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockPVZRepo)
	mockMetrics := new(mockMetrics)
	svc := NewPVZService(mockRepo, nil, nil, mockMetrics)

	t.Run("invalid city", func(t *testing.T) {
		_, err := svc.CreatePVZ(ctx, oapi.PostPvzJSONRequestBody{City: "X"})
		require.ErrorIs(t, err, pvz_errors.ErrInvalidPVZCity)
	})

	t.Run("insert error", func(t *testing.T) {
		req := oapi.PostPvzJSONRequestBody{City: oapi.Москва}
		mockRepo.
			On("InsertPVZ", mock.Anything, req.City, mock.Anything).
			Return(oapi.PVZ{}, errors.New("db"))
		_, err := svc.CreatePVZ(ctx, req)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		req := oapi.PostPvzJSONRequestBody{City: oapi.Казань}
		returned := oapi.PVZ{Id: uuidPtr(uuid.New()), City: req.City}
		mockRepo.
			On("InsertPVZ", mock.Anything, req.City, mock.Anything).
			Return(returned, nil)
		mockMetrics.
			On("SendBusinessMetricsUpdate", metrics.MetricsUpdate{PvzCreatedDelta: 1}).
			Return()

		result, err := svc.CreatePVZ(ctx, req)
		require.NoError(t, err)
		require.Equal(t, returned, result)

		mockRepo.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

func TestGetPVZ(t *testing.T) {
	ctx := context.Background()

	t.Run("select error", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, time.Time{}, mock.Anything, 10, 0).
			Return([]oapi.PVZ(nil), errors.New("fail"))

		_, err := svc.GetPVZ(ctx, oapi.GetPvzParams{})
		require.Error(t, err)
		mockPVZ.AssertExpectations(t)
	})

	t.Run("no receptions", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		now := time.Now()
		pvzList := []oapi.PVZ{{Id: uuidPtr(uuid.New()), City: oapi.Москва, RegistrationDate: &now}}
		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, time.Time{}, mock.Anything, 10, 0).
			Return(pvzList, nil)
		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, *pvzList[0].Id).
			Return([]dto.Reception{}, nil)

		result, err := svc.GetPVZ(ctx, oapi.GetPvzParams{})
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Len(t, result[0].Receptions, 0)

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
	})

	t.Run("with products error", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		now := time.Now()
		id := uuid.New()
		pvz := oapi.PVZ{Id: &id, City: oapi.Москва, RegistrationDate: &now}
		recs := []dto.Reception{{Id: uuidPtr(uuid.New()), PvzId: id}}

		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, time.Time{}, mock.Anything, 10, 0).
			Return([]oapi.PVZ{pvz}, nil)
		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, id).
			Return(recs, nil)
		ids := []*uuid.UUID{recs[0].Id}
		mockProd.
			On("GetProductsByReceptionIDs", mock.Anything, ids).
			Return([]oapi.Product(nil), errors.New("fail"))

		_, err := svc.GetPVZ(ctx, oapi.GetPvzParams{})
		require.Error(t, err)

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
		mockProd.AssertExpectations(t)
	})
}

func TestGetAllPVZs(t *testing.T) {
	ctx := context.Background()

	t.Run("error", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		svc := NewPVZService(mockPVZ, nil, nil, nil)

		mockPVZ.
			On("SelectAllPVZs", mock.Anything).
			Return([]*proto.PVZ(nil), errors.New("x")).
			Once()

		_, err := svc.GetAllPVZs(ctx)
		require.Error(t, err)
		mockPVZ.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		svc := NewPVZService(mockPVZ, nil, nil, nil)

		list := []*proto.PVZ{{Id: uuid.New().String()}}
		mockPVZ.
			On("SelectAllPVZs", mock.Anything).
			Return(list, nil).
			Once()

		res, err := svc.GetAllPVZs(ctx)
		require.NoError(t, err)
		require.Equal(t, list, res)
		mockPVZ.AssertExpectations(t)
	})
}
