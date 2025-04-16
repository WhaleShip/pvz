package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
)

type ReceptionRepository interface {
	CreateReception(ctx context.Context, req gen.PostReceptionsJSONRequestBody) (gen.Reception, error)
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (gen.Reception, error)
}

type receptionRepository struct {
	db *pgxpool.Pool
}

func NewReceptionRepository(dbConn *pgxpool.Pool) ReceptionRepository {
	return &receptionRepository{db: dbConn}
}

func (r *receptionRepository) CreateReception(
	ctx context.Context,
	req gen.PostReceptionsJSONRequestBody) (gen.Reception, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return gen.Reception{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	newReceptionID := uuid.New()
	now := time.Now()

	var insertedID uuid.UUID
	err = tx.QueryRow(ctx, QueryInsertReception,
		req.PvzId,
		newReceptionID,
		now,
	).Scan(&insertedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.Reception{}, pvz_errors.ErrPVZNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return gen.Reception{}, pvz_errors.ErrPVZNotFound
			case "23505":
				if pgErr.ConstraintName == "idx_unique_open_reception" {
					return gen.Reception{}, pvz_errors.ErrOpenReceptionExists
				}
			}
		}
		return gen.Reception{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return gen.Reception{}, err
	}

	return gen.Reception{
		Id:       &insertedID,
		DateTime: now,
		PvzId:    req.PvzId,
		Status:   gen.ReceptionStatus("in_progress"),
	}, nil
}

func (r *receptionRepository) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (gen.Reception, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return gen.Reception{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var receptionID uuid.UUID
	err = tx.QueryRow(ctx, QueryCloseActiveReception, pvzID).Scan(&receptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.Reception{}, pvz_errors.ErrCloseReceptionFailed
		}
		return gen.Reception{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return gen.Reception{}, err
	}

	now := time.Now()
	return gen.Reception{
		Id:       &receptionID,
		DateTime: now,
		PvzId:    pvzID,
		Status:   gen.ReceptionStatus("close"),
	}, nil
}
