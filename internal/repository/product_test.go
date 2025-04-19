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
)

func TestInsertProduct(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewProductRepository(db)

	ctx := context.Background()
	pvzID := uuid.New()
	productID := uuid.New()
	now := time.Now()
	typ := "standard"
	newRecv := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertProduct).
			WithArgs(pvzID, productID, now, typ).
			WillReturnRows(pgxmock.NewRows([]string{"reception_id"}).AddRow(newRecv))
		mockPool.ExpectCommit()

		got, err := repo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.NoError(t, err)
		require.Equal(t, newRecv, got)
	})

	t.Run("no open reception or pvz", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertProduct).
			WithArgs(pvzID, productID, now, typ).
			WillReturnError(db.ErrNoRows())

		_, err := repo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.ErrorIs(t, err, pvz_errors.ErrNoOpenRecetionOrPvz)
	})

	t.Run("invalid product constraint", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23514"}
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertProduct).
			WithArgs(pvzID, productID, now, typ).
			WillReturnError(pgErr)

		_, err := repo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.ErrorIs(t, err, pvz_errors.ErrInvalidProduct)
	})

	t.Run("other db error", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertProduct).
			WithArgs(pvzID, productID, now, typ).
			WillReturnError(errors.New("some db error"))

		_, err := repo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.Error(t, err)
	})

	t.Run("commit error", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(QueryInsertProduct).
			WithArgs(pvzID, productID, now, typ).
			WillReturnRows(pgxmock.NewRows([]string{"reception_id"}).AddRow(newRecv))
		mockPool.ExpectCommit().WillReturnError(errors.New("commit failed"))

		_, err := repo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.Error(t, err)
	})

	t.Run("begin error", func(t *testing.T) {
		badPool, _ := pgxmock.NewPool()
		defer badPool.Close()
		badDb := &database.PgxMockAdapter{Pool: badPool}
		badRepo := NewProductRepository(badDb)

		badPool.ExpectBegin().WillReturnError(errors.New("begin failed"))
		_, err := badRepo.InsertProduct(ctx, pvzID, productID, now, typ)
		require.Error(t, err)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewProductRepository(db)

	pvzID := uuid.New()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectExec(QueryDeleteLastProduct).
			WithArgs(pvzID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mockPool.ExpectCommit()

		require.NoError(t, repo.DeleteLastProduct(ctx, pvzID))
	})

	t.Run("nothing to delete", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectExec(QueryDeleteLastProduct).
			WithArgs(pvzID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.DeleteLastProduct(ctx, pvzID)
		require.ErrorIs(t, err, pvz_errors.ErrDeletingProduct)
	})

	t.Run("exec error", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectExec(QueryDeleteLastProduct).
			WithArgs(pvzID).
			WillReturnError(errors.New("exec failed"))

		err := repo.DeleteLastProduct(ctx, pvzID)
		require.Error(t, err)
	})

	t.Run("commit error", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.
			ExpectExec(QueryDeleteLastProduct).
			WithArgs(pvzID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mockPool.ExpectCommit().WillReturnError(errors.New("commit failed"))

		err := repo.DeleteLastProduct(ctx, pvzID)
		require.Error(t, err)
	})

	t.Run("begin error", func(t *testing.T) {
		badPool, _ := pgxmock.NewPool()
		defer badPool.Close()
		badDb := &database.PgxMockAdapter{Pool: badPool}
		badRepo := NewProductRepository(badDb)

		badPool.ExpectBegin().WillReturnError(errors.New("begin failed"))
		err := badRepo.DeleteLastProduct(ctx, pvzID)
		require.Error(t, err)
	})
}

func TestGetProductsByReceptionIDs(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewProductRepository(db)

	ctx := context.Background()
	ids := []*uuid.UUID{uuidPtr(uuid.New()), uuidPtr(uuid.New())}

	t.Run("success multiple products", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "reception_id", "date_time", "type"}).
			AddRow(uuid.New(), *ids[0], time.Now(), "A").
			AddRow(uuid.New(), *ids[1], time.Now(), "B")
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnRows(rows)

		prods, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.NoError(t, err)
		require.Len(t, prods, 2)
	})

	t.Run("query error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnError(errors.New("query failed"))

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.Error(t, err)
	})

	t.Run("query no rows", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnError(db.ErrNoRows())

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.ErrorIs(t, err, pvz_errors.ErrSelectProductsFailed)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "reception_id", "date_time", "type"}).
			AddRow("bad-uuid", *ids[0], time.Now(), "A")
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnRows(rows)

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.Error(t, err)
	})

	t.Run("scan no rows", func(t *testing.T) {
		validID := uuid.New()
		rows := pgxmock.NewRows([]string{"id", "reception_id", "date_time", "type"}).
			AddRow(validID, *ids[0], time.Now(), "A").
			RowError(0, db.ErrNoRows())
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnRows(rows)

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.ErrorIs(t, err, pvz_errors.ErrSelectProductsFailed)
	})

	t.Run("rows error after next", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "reception_id", "date_time", "type"}).
			RowError(0, errors.New("row fail"))
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnRows(rows)

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.Error(t, err)
	})

	t.Run("rows Err no rows", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "reception_id", "date_time", "type"}).
			RowError(0, db.ErrNoRows())
		mockPool.
			ExpectQuery(QueryGetProductsByReceptions).
			WithArgs(ids).
			WillReturnRows(rows)

		_, err := repo.GetProductsByReceptionIDs(ctx, ids)
		require.ErrorIs(t, err, pvz_errors.ErrSelectProductsFailed)
	})
}

func uuidPtr(u uuid.UUID) *uuid.UUID { return &u }
