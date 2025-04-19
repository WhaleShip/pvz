package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/metrics"
)

type mockProductRepo struct{ mock.Mock }

func (m *mockProductRepo) InsertProduct(
	ctx context.Context, pvzID,
	productID uuid.UUID,
	dateTime time.Time,
	productType string) (uuid.UUID, error) {
	args := m.Called(ctx, pvzID, productID, dateTime, productType)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockProductRepo) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return m.Called(ctx, pvzID).Error(0)
}

func TestAddProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockProductRepo)
	mockMetrics := new(mockMetrics)
	svc := NewProductService(mockRepo, mockMetrics)

	t.Run("insert error", func(t *testing.T) {
		pvzID := uuid.New()
		req := oapi.PostProductsJSONRequestBody{
			PvzId: pvzID,
			Type:  oapi.PostProductsJSONBodyType("T"),
		}
		mockRepo.
			On("InsertProduct",
				mock.Anything,
				pvzID,
				mock.AnythingOfType("uuid.UUID"),
				mock.AnythingOfType("time.Time"),
				string(req.Type),
			).
			Return(uuid.Nil, errors.New("fail"))

		_, err := svc.AddProduct(ctx, req)
		require.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New()
		req := oapi.PostProductsJSONRequestBody{
			PvzId: pvzID,
			Type:  oapi.PostProductsJSONBodyType("T"),
		}

		expectedReceptionID := uuid.New()

		var capturedProductID uuid.UUID
		var capturedTime time.Time

		mockRepo.
			On("InsertProduct",
				mock.Anything,
				pvzID,
				mock.MatchedBy(func(id uuid.UUID) bool {
					capturedProductID = id
					return true
				}),
				mock.MatchedBy(func(tm time.Time) bool {
					capturedTime = tm
					return true
				}),
				string(req.Type),
			).
			Return(expectedReceptionID, nil)

		mockMetrics.
			On("SendBusinessMetricsUpdate", metrics.MetricsUpdate{ProductsAddedDelta: 1}).
			Return()

		prod, err := svc.AddProduct(ctx, req)
		require.NoError(t, err)

		require.Equal(t, &capturedProductID, prod.Id)

		require.Equal(t, expectedReceptionID, prod.ReceptionId)

		require.Equal(t, oapi.ProductType(req.Type), prod.Type)

		require.WithinDuration(t, capturedTime, *prod.DateTime, time.Millisecond*10)

		mockRepo.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
	t.Run("metrics nil", func(t *testing.T) {
		mockRepo := new(mockProductRepo)
		svc := NewProductService(mockRepo, nil)

		pvzID := uuid.New()
		req := oapi.PostProductsJSONRequestBody{PvzId: pvzID, Type: "T"}
		expectedReceptionID := uuid.New()

		var capturedProductID uuid.UUID

		mockRepo.
			On("InsertProduct",
				mock.Anything,
				pvzID,
				mock.MatchedBy(func(id uuid.UUID) bool {
					capturedProductID = id
					return true
				}),
				mock.MatchedBy(func(tm time.Time) bool {
					return true
				}),
				string(req.Type),
			).
			Return(expectedReceptionID, nil)

		prod, err := svc.AddProduct(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, &capturedProductID, prod.Id)
		require.Equal(t, expectedReceptionID, prod.ReceptionId)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockProductRepo)
	svc := NewProductService(mockRepo, nil)

	t.Run("delete error", func(t *testing.T) {
		pvzID := uuid.New()
		mockRepo.
			On("DeleteLastProduct", mock.Anything, pvzID).
			Return(errors.New("nope"))

		err := svc.DeleteLastProduct(ctx, pvzID)
		require.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		pvzID := uuid.New()
		mockRepo.
			On("DeleteLastProduct", mock.Anything, pvzID).
			Return(nil)

		err := svc.DeleteLastProduct(ctx, pvzID)
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}
