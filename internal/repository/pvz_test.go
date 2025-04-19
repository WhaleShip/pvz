package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	pvz_errors "github.com/whaleship/pvz/internal/errors"

	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

func TestInsertPVZ(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewPVZRepository(db)

	ctx := context.Background()
	city := oapi.Москва
	reg := time.Now()

	t.Run("success", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryInsertPVZ).
			WithArgs(
				pgxmock.AnyArg(),
				string(city),
				pgxmock.AnyArg(),
			).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "city", "registration_date"}).
					AddRow(uuid.New(), string(city), reg),
			)

		pvz, err := repo.InsertPVZ(ctx, city, reg)
		require.NoError(t, err)
		require.Equal(t, string(city), string(pvz.City))
	})

	t.Run("no rows", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryInsertPVZ).
			WithArgs(
				pgxmock.AnyArg(),
				string(city),
				pgxmock.AnyArg(),
			).
			WillReturnError(db.ErrNoRows())

		_, err := repo.InsertPVZ(ctx, city, reg)
		require.ErrorIs(t, err, pvz_errors.ErrInsertPVZFailed)
	})

	t.Run("other error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryInsertPVZ).
			WithArgs(
				pgxmock.AnyArg(),
				string(city),
				pgxmock.AnyArg(),
			).
			WillReturnError(errors.New("db"))

		_, err := repo.InsertPVZ(ctx, city, reg)
		require.Error(t, err)
	})
	t.Run("scan error", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow("not-a-uuid", string(city), reg)
		mockPool.
			ExpectQuery(QueryInsertPVZ).
			WithArgs(
				pgxmock.AnyArg(),
				string(city),
				pgxmock.AnyArg(),
			).
			WillReturnRows(rows)

		_, err := repo.InsertPVZ(ctx, city, reg)
		require.Error(t, err)
	})

	t.Run("query error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryInsertPVZ).
			WillReturnError(errors.New("some failure"))

		_, err := repo.InsertPVZ(ctx, city, reg)
		require.Error(t, err)
	})
}

func TestSelectPVZByOpenReceptions(t *testing.T) {
	mockPool, _ := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewPVZRepository(db)

	ctx := context.Background()
	start, end := time.Now(), time.Now().Add(time.Hour)

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(uuid.New(), "X", start).
			AddRow(uuid.New(), "Y", end)
		mockPool.
			ExpectQuery(QuerySelectPVZByOpenReceptions).
			WithArgs(start, end, 10, 0).
			WillReturnRows(rows)

		list, err := repo.SelectPVZByOpenReceptions(ctx, start, end, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("query fail", func(t *testing.T) {
		mockPool.
			ExpectQuery(QuerySelectPVZByOpenReceptions).
			WillReturnError(fmt.Errorf("err"))

		_, err := repo.SelectPVZByOpenReceptions(ctx, start, end, 1, 0)
		require.Error(t, err)
	})
	t.Run("scan error", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow("bad-uuid", "X", start)
		mockPool.
			ExpectQuery(QuerySelectPVZByOpenReceptions).
			WithArgs(start, end, 10, 0).
			WillReturnRows(rows)

		_, err := repo.SelectPVZByOpenReceptions(ctx, start, end, 10, 0)
		require.Error(t, err)
	})

	t.Run("rows.Err", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			RowError(0, errors.New("row failure"))
		mockPool.
			ExpectQuery(QuerySelectPVZByOpenReceptions).
			WithArgs(start, end, 5, 1).
			WillReturnRows(rows)

		_, err := repo.SelectPVZByOpenReceptions(ctx, start, end, 5, 1)
		require.Error(t, err)
	})
}

func TestSelectAllPVZs(t *testing.T) {
	mockPool, _ := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewPVZRepository(db)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(uuid.New().String(), "A", time.Now()).
			AddRow(uuid.New().String(), "B", time.Now())
		mockPool.
			ExpectQuery(QuerySelectAllPVZs).
			WillReturnRows(rows)

		out, err := repo.SelectAllPVZs(ctx)
		require.NoError(t, err)
		require.Len(t, out, 2)
	})

	t.Run("rows.Err", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			RowError(0, errors.New("bad"))
		mockPool.
			ExpectQuery(QuerySelectAllPVZs).
			WillReturnRows(rows)

		_, err := repo.SelectAllPVZs(ctx)
		require.Error(t, err)
	})
	t.Run("query error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QuerySelectAllPVZs).
			WillReturnError(errors.New("fatal"))

		_, err := repo.SelectAllPVZs(ctx)
		require.Error(t, err)
	})
	t.Run("scan error", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
			AddRow(uuid.New().String(), "C", "not-a-timestamp")
		mockPool.
			ExpectQuery(QuerySelectAllPVZs).
			WillReturnRows(rows)

		_, err := repo.SelectAllPVZs(ctx)
		require.Error(t, err)
	})
}
