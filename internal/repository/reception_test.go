package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/database"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

func TestCreateReception(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewReceptionRepository(db)

	ctx := context.Background()
	req := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}

	t.Run("success", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertReception).
			WithArgs(
				req.PvzId,
				pgxmock.AnyArg(),
				pgxmock.AnyArg(),
			).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mockPool.ExpectCommit()

		got, err := repo.CreateReception(ctx, req)
		require.NoError(t, err)
		require.Equal(t, oapi.ReceptionStatus("in_progress"), got.Status)
	})

	t.Run("pvz not found", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertReception).
			WithArgs(
				req.PvzId,
				pgxmock.AnyArg(),
				pgxmock.AnyArg(),
			).
			WillReturnError(db.ErrNoRows())
		_, err := repo.CreateReception(ctx, req)
		require.ErrorIs(t, err, pvz_errors.ErrPVZNotFound)
	})

	t.Run("unique constraint", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertReception).
			WithArgs(
				req.PvzId,
				pgxmock.AnyArg(),
				pgxmock.AnyArg(),
			).
			WillReturnError(&pgconn.PgError{
				Code:           "23505",
				ConstraintName: "idx_unique_open_reception",
			})
		_, err := repo.CreateReception(ctx, req)
		require.ErrorIs(t, err, pvz_errors.ErrOpenReceptionExists)
	})

	t.Run("other db error", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertReception).
			WithArgs(
				req.PvzId,
				pgxmock.AnyArg(),
				pgxmock.AnyArg(),
			).
			WillReturnError(errors.New("some db failure"))
		_, err := repo.CreateReception(ctx, req)
		require.Error(t, err)
	})
}

func TestCloseLastReception(t *testing.T) {
	mockPool, err := pgxmock.NewPool(
		pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
	)
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewReceptionRepository(db)

	ctx := context.Background()
	pvzID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryCloseActiveReception).
			WithArgs(pvzID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "open_time"}).
					AddRow(uuid.New(), time.Now()),
			)
		mockPool.ExpectCommit()

		got, err := repo.CloseLastReception(ctx, pvzID)
		require.NoError(t, err)
		require.Equal(t, oapi.ReceptionStatus("close"), got.Status)
	})

	t.Run("nothing to close", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryCloseActiveReception).
			WithArgs(pvzID).
			WillReturnError(db.ErrNoRows())
		_, err := repo.CloseLastReception(ctx, pvzID)
		require.ErrorIs(t, err, pvz_errors.ErrCloseReceptionFailed)
	})
}

func TestGetReceptionsByPVZ(t *testing.T) {
	mockPool, _ := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewReceptionRepository(db)

	ctx := context.Background()
	pvzID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "pvz_id", "open_time", "status"}).
			AddRow(uuid.New(), pvzID, time.Now(), "in_progress")
		mockPool.
			ExpectQuery(QueryGetReceptionsByPVZs).
			WithArgs(pvzID).
			WillReturnRows(rows)

		out, err := repo.GetReceptionsByPVZ(ctx, pvzID)
		require.NoError(t, err)
		require.Len(t, out, 1)
		require.Equal(t, oapi.InProgress, out[0].Status)
	})

	t.Run("query error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryGetReceptionsByPVZs).
			WillReturnError(errors.New("fail"))

		_, err := repo.GetReceptionsByPVZ(ctx, pvzID)
		require.Error(t, err)
	})
}
