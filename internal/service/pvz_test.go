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
	t.Run("metrics nil", func(t *testing.T) {
		mockRepo := new(mockPVZRepo)
		svc := NewPVZService(mockRepo, nil, nil, nil)

		req := oapi.PostPvzJSONRequestBody{City: oapi.Москва}
		returned := oapi.PVZ{Id: uuidPtr(uuid.New()), City: req.City}

		mockRepo.
			On("InsertPVZ", mock.Anything, req.City, mock.Anything).
			Return(returned, nil).
			Once()

		res, err := svc.CreatePVZ(ctx, req)
		require.NoError(t, err)
		require.Equal(t, returned, res)

		mockRepo.AssertExpectations(t)
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

	t.Run("success with data", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		now := time.Now()
		pvzID := uuid.New()
		pvz := oapi.PVZ{Id: &pvzID, City: oapi.Москва, RegistrationDate: &now}
		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, mock.Anything, mock.Anything, 10, 0).
			Return([]oapi.PVZ{pvz}, nil)

		rec1 := dto.Reception{Id: uuidPtr(uuid.New()), PvzId: pvzID, DateTime: now, Status: oapi.InProgress}
		rec2 := dto.Reception{Id: uuidPtr(uuid.New()), PvzId: pvzID, DateTime: now.Add(time.Hour), Status: oapi.Close}
		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, pvzID).
			Return([]dto.Reception{rec1, rec2}, nil)

		prod1 := oapi.Product{Id: uuidPtr(uuid.New()), ReceptionId: *rec1.Id, DateTime: &now, Type: oapi.ProductType("X")}
		prod2 := oapi.Product{Id: uuidPtr(uuid.New()), ReceptionId: *rec2.Id, DateTime: &now, Type: oapi.ProductType("Y")}
		mockProd.
			On("GetProductsByReceptionIDs", mock.Anything, []*uuid.UUID{rec1.Id, rec2.Id}).
			Return([]oapi.Product{prod1, prod2}, nil)

		out, err := svc.GetPVZ(context.Background(), oapi.GetPvzParams{})
		require.NoError(t, err)
		require.Len(t, out, 1)
		entry := out[0]
		require.Equal(t, pvz, entry.Pvz)
		require.Len(t, entry.Receptions, 2)
		require.Equal(t, prod1, entry.Receptions[0].Products[0])
		require.Equal(t, prod2, entry.Receptions[1].Products[0])

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
		mockProd.AssertExpectations(t)
	})

	t.Run("with start and end dates", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		params := oapi.GetPvzParams{
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		pvzList := []oapi.PVZ{{Id: uuidPtr(uuid.New()), City: oapi.Москва, RegistrationDate: &endDate}}
		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, startDate, endDate, 10, 0).
			Return(pvzList, nil)
		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, *pvzList[0].Id).
			Return([]dto.Reception{}, nil)

		out, err := svc.GetPVZ(ctx, params)
		require.NoError(t, err)
		require.Len(t, out, 1)
		require.Equal(t, pvzList[0], out[0].Pvz)
		require.Len(t, out[0].Receptions, 0)

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
	})

	t.Run("pagination params", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		svc := NewPVZService(mockPVZ, nil, nil, nil)

		page, limit := 2, 5
		params := oapi.GetPvzParams{Page: &page, Limit: &limit}
		mockPVZ.
			On("SelectPVZByOpenReceptions",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				limit,
				(page-1)*limit,
			).
			Return([]oapi.PVZ{}, nil)

		out, err := svc.GetPVZ(context.Background(), params)
		require.NoError(t, err)
		require.Len(t, out, 0)

		mockPVZ.AssertExpectations(t)
	})
	t.Run("skip receptions not belonging to PVZ", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		now := time.Now()
		pvzID1 := uuid.New()
		pvzID2 := uuid.New()
		pvz1 := oapi.PVZ{Id: &pvzID1, City: oapi.Москва, RegistrationDate: &now}
		pvz2 := oapi.PVZ{Id: &pvzID2, City: oapi.Казань, RegistrationDate: &now}

		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, mock.Anything, mock.Anything, 10, 0).
			Return([]oapi.PVZ{pvz1, pvz2}, nil)

		rec1 := dto.Reception{Id: uuidPtr(uuid.New()), PvzId: pvzID1, DateTime: now, Status: oapi.InProgress}
		rec2 := dto.Reception{Id: uuidPtr(uuid.New()), PvzId: pvzID2, DateTime: now, Status: oapi.InProgress}
		rec3 := dto.Reception{Id: uuidPtr(uuid.New()), PvzId: pvzID1, DateTime: now, Status: oapi.Close}

		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, pvzID1).
			Return([]dto.Reception{rec1, rec3}, nil)
		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, pvzID2).
			Return([]dto.Reception{rec2}, nil)

		prod1 := oapi.Product{Id: uuidPtr(uuid.New()), ReceptionId: *rec1.Id, DateTime: &now, Type: oapi.ProductType("X")}
		prod2 := oapi.Product{Id: uuidPtr(uuid.New()), ReceptionId: *rec2.Id, DateTime: &now, Type: oapi.ProductType("Y")}
		prod3 := oapi.Product{Id: uuidPtr(uuid.New()), ReceptionId: *rec3.Id, DateTime: &now, Type: oapi.ProductType("Z")}

		mockProd.
			On("GetProductsByReceptionIDs",
				mock.Anything, mock.MatchedBy(uuidSliceMatcher([]*uuid.UUID{rec1.Id, rec2.Id, rec3.Id}))).
			Return([]oapi.Product{prod1, prod2, prod3}, nil)

		out, err := svc.GetPVZ(ctx, oapi.GetPvzParams{})
		require.NoError(t, err)
		require.Len(t, out, 2)

		pvz1Entry := out[0]
		require.Equal(t, pvz1, pvz1Entry.Pvz)
		require.Len(t, pvz1Entry.Receptions, 2)
		require.Equal(t, rec1, pvz1Entry.Receptions[0].Reception)
		require.Equal(t, prod1, pvz1Entry.Receptions[0].Products[0])
		require.Equal(t, rec3, pvz1Entry.Receptions[1].Reception)
		require.Equal(t, prod3, pvz1Entry.Receptions[1].Products[0])

		pvz2Entry := out[1]
		require.Equal(t, pvz2, pvz2Entry.Pvz)
		require.Len(t, pvz2Entry.Receptions, 1)
		require.Equal(t, rec2, pvz2Entry.Receptions[0].Reception)
		require.Equal(t, prod2, pvz2Entry.Receptions[0].Products[0])

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
		mockProd.AssertExpectations(t)
	})
	t.Run("error getting receptions", func(t *testing.T) {
		mockPVZ := new(mockPVZRepo)
		mockRec := new(mockReceptionReader)
		mockProd := new(mockProductReader)
		svc := NewPVZService(mockPVZ, mockRec, mockProd, nil)

		now := time.Now()
		pvzID := uuid.New()
		pvz := oapi.PVZ{Id: &pvzID, City: oapi.Москва, RegistrationDate: &now}

		mockPVZ.
			On("SelectPVZByOpenReceptions", mock.Anything, mock.Anything, mock.Anything, 10, 0).
			Return([]oapi.PVZ{pvz}, nil)

		mockRec.
			On("GetReceptionsByPVZ", mock.Anything, pvzID).
			Return([]dto.Reception{}, errors.New("db error"))

		_, err := svc.GetPVZ(ctx, oapi.GetPvzParams{})
		require.Error(t, err)
		require.ErrorIs(t, err, pvz_errors.ErrSelectReceptionsFailed)
		require.Contains(t, err.Error(), "db error")

		mockPVZ.AssertExpectations(t)
		mockRec.AssertExpectations(t)
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

func uuidSliceMatcher(expected []*uuid.UUID) func([]*uuid.UUID) bool {
	return func(actual []*uuid.UUID) bool {
		if len(actual) != len(expected) {
			return false
		}
		actualSet := make(map[uuid.UUID]bool)
		for _, id := range actual {
			actualSet[*id] = true
		}
		for _, id := range expected {
			if !actualSet[*id] {
				return false
			}
		}
		return true
	}
}
