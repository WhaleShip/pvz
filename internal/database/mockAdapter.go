package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
)

type PgxMockAdapter struct {
	Pool pgxmock.PgxPoolIface
}

func (a *PgxMockAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	return a.Pool.Begin(ctx)
}

func (a *PgxMockAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return a.Pool.Exec(ctx, sql, args...)
}

func (a *PgxMockAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return a.Pool.Query(ctx, sql, args...)
}

func (a *PgxMockAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return a.Pool.QueryRow(ctx, sql, args...)
}

func (a *PgxMockAdapter) ErrNoRows() error {
	return pgx.ErrNoRows
}
