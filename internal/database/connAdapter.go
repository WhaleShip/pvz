package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxPoolAdapter struct {
	Pool *pgxpool.Pool
}

func (a *PgxPoolAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	return a.Pool.Begin(ctx)
}

func (a *PgxPoolAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return a.Pool.Exec(ctx, sql, args...)
}

func (a *PgxPoolAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return a.Pool.Query(ctx, sql, args...)
}

func (a *PgxPoolAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return a.Pool.QueryRow(ctx, sql, args...)
}

func (a *PgxPoolAdapter) ErrNoRows() error {
	return pgx.ErrNoRows
}
