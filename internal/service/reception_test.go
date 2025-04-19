package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/metrics"
)

type mockReceptionWriter struct{ mock.Mock }

func (m *mockReceptionWriter) CreateReception(
	ctx context.Context,
	req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(oapi.Reception), args.Error(1)
}

func (m *mockReceptionWriter) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (oapi.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(oapi.Reception), args.Error(1)
}

func TestCreateReception(t *testing.T) {
	mockRepo := new(mockReceptionWriter)
	mockMetrics := new(mockMetrics)
	svc := NewReceptionService(mockRepo, mockMetrics)

	t.Run("error", func(t *testing.T) {
		req := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}
		mockRepo.
			On("CreateReception", mock.Anything, req).
			Return(oapi.Reception{}, errors.New("fail"))
		_, err := svc.CreateReception(req)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		req := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}
		ret := oapi.Reception{Id: uuidPtr(uuid.New())}
		mockRepo.
			On("CreateReception", mock.Anything, req).
			Return(ret, nil)
		mockMetrics.
			On("SendBusinessMetricsUpdate", metrics.MetricsUpdate{ReceptionsCreatedDelta: 1}).
			Return()
		res, err := svc.CreateReception(req)
		require.NoError(t, err)
		require.Equal(t, ret, res)
		mockRepo.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
	t.Run("metrics nil", func(t *testing.T) {
		mockRepo := new(mockReceptionWriter)
		svc := NewReceptionService(mockRepo, nil)

		req := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}
		expected := oapi.Reception{Id: uuidPtr(uuid.New())}

		mockRepo.
			On("CreateReception", mock.Anything, req).
			Return(expected, nil)

		out, err := svc.CreateReception(req)
		require.NoError(t, err)
		require.Equal(t, expected, out)
		mockRepo.AssertExpectations(t)
	})
}

func TestCloseLastReception(t *testing.T) {
	mockRepo := new(mockReceptionWriter)
	svc := NewReceptionService(mockRepo, nil)

	t.Run("error", func(t *testing.T) {
		id := uuid.New()
		mockRepo.
			On("CloseLastReception", mock.Anything, id).
			Return(oapi.Reception{}, errors.New("no"))
		_, err := svc.CloseLastReception(id)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		ret := oapi.Reception{Id: uuidPtr(uuid.New())}
		mockRepo.
			On("CloseLastReception", mock.Anything, id).
			Return(ret, nil)
		res, err := svc.CloseLastReception(id)
		require.NoError(t, err)
		require.Equal(t, ret, res)
		mockRepo.AssertExpectations(t)
	})
}
