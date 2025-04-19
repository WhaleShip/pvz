package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/database"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
)

func TestInsertUser(t *testing.T) {
	mockPool, _ := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewUserRepository(db)

	ctx := context.Background()
	id := uuid.New()
	email, pass, role := "a@b.c", "pwd", "admin"

	t.Run("success", func(t *testing.T) {
		mockPool.
			ExpectExec(QueryInsertUser).
			WithArgs(id, email, pass, role).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		require.NoError(t, repo.InsertUser(ctx, id, email, pass, role))
	})

	t.Run("already exists", func(t *testing.T) {
		mockPool.
			ExpectExec(QueryInsertUser).
			WithArgs(id, email, pass, role).
			WillReturnResult(pgxmock.NewResult("INSERT", 0))

		err := repo.InsertUser(ctx, id, email, pass, role)
		require.ErrorIs(t, err, pvz_errors.ErrUserAlreadyExists)
	})

	t.Run("exec error", func(t *testing.T) {
		mockPool.
			ExpectExec(QueryInsertUser).
			WillReturnError(errors.New("boom"))

		err := repo.InsertUser(ctx, id, email, pass, role)
		require.Error(t, err)
	})
	t.Run("exec pg error", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505"}
		mockPool.
			ExpectExec(QueryInsertUser).
			WithArgs(id, email, pass, role).
			WillReturnError(pgErr)

		err := repo.InsertUser(ctx, id, email, pass, role)
		require.Error(t, err)
	})
}

func TestGetUserByEmail(t *testing.T) {
	mockPool, _ := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	defer mockPool.Close()

	db := &database.PgxMockAdapter{Pool: mockPool}
	repo := NewUserRepository(db)

	ctx := context.Background()
	email := "x@y.z"
	userID := uuid.New()
	pass := "hash"
	role := "user"

	t.Run("success", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryUserByEmail).
			WithArgs(email).
			WillReturnRows(pgxmock.NewRows([]string{"id", "password", "role"}).
				AddRow(userID, pass, role))

		id, pw, r, err := repo.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		require.Equal(t, userID, id)
		require.Equal(t, pass, pw)
		require.Equal(t, role, r)
	})

	t.Run("not found", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryUserByEmail).
			WithArgs(email).
			WillReturnError(db.ErrNoRows())

		_, _, _, err := repo.GetUserByEmail(ctx, email)
		require.ErrorIs(t, err, pvz_errors.ErrUserNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryUserByEmail).
			WithArgs(email).
			WillReturnError(errors.New("db"))

		_, _, _, err := repo.GetUserByEmail(ctx, email)
		require.Error(t, err)
	})
	t.Run("scan error", func(t *testing.T) {
		mockPool.
			ExpectQuery(QueryUserByEmail).
			WithArgs(email).
			WillReturnRows(pgxmock.NewRows([]string{"id", "password", "role"}).AddRow("not-uuid", pass, role))

		_, _, _, err := repo.GetUserByEmail(ctx, email)
		require.Error(t, err)
	})
}
