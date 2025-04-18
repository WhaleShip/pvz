package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/dto"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type receptionRepository struct {
	db database.PgxIface
}

func NewReceptionRepository(dbConn database.PgxIface) *receptionRepository {
	return &receptionRepository{db: dbConn}
}

func (r *receptionRepository) CreateReception(
	ctx context.Context,
	req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return oapi.Reception{}, err
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
		if errors.Is(err, r.db.ErrNoRows()) {
			return oapi.Reception{}, pvz_errors.ErrPVZNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return oapi.Reception{}, pvz_errors.ErrPVZNotFound
			case "23505":
				if pgErr.ConstraintName == "idx_unique_open_reception" {
					return oapi.Reception{}, pvz_errors.ErrOpenReceptionExists
				}
			}
		}
		return oapi.Reception{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return oapi.Reception{}, err
	}

	return oapi.Reception{
		Id:       &insertedID,
		DateTime: now,
		PvzId:    req.PvzId,
		Status:   oapi.ReceptionStatus("in_progress"),
	}, nil
}

func (r *receptionRepository) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (oapi.Reception, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return oapi.Reception{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var (
		receptionID uuid.UUID
		openTime    time.Time
	)
	err = tx.QueryRow(ctx, QueryCloseActiveReception, pvzID).
		Scan(&receptionID, &openTime)
	if err != nil {
		if errors.Is(err, r.db.ErrNoRows()) {
			return oapi.Reception{}, pvz_errors.ErrCloseReceptionFailed
		}
		return oapi.Reception{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return oapi.Reception{}, err
	}

	return oapi.Reception{
		Id:       &receptionID,
		PvzId:    pvzID,
		DateTime: openTime,
		Status:   oapi.ReceptionStatus("close"),
	}, nil
}

func (r *receptionRepository) GetReceptionsByPVZ(ctx context.Context, pvzID uuid.UUID) ([]dto.Reception, error) {
	rows, err := r.db.Query(ctx, QueryGetReceptionsByPVZs, pvzID)
	if err != nil {
		if errors.Is(err, r.db.ErrNoRows()) {
			return nil, pvz_errors.ErrSelectReceptionsFailed
		}
		return nil, err
	}
	defer rows.Close()

	var receptions []dto.Reception
	for rows.Next() {
		var (
			id        uuid.UUID
			pvzId     uuid.UUID
			openTime  time.Time
			closeTime *time.Time
			status    string
		)
		if err := rows.Scan(&id, &pvzId, &openTime, &status); err != nil {
			if errors.Is(err, r.db.ErrNoRows()) {
				return nil, pvz_errors.ErrSelectReceptionsFailed
			}
			return nil, err
		}
		receptions = append(receptions, dto.Reception{
			Id:            &id,
			PvzId:         pvzId,
			DateTime:      openTime,
			CloseDateTime: closeTime,
			Status:        oapi.ReceptionStatus(status),
		})
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, r.db.ErrNoRows()) {
			return nil, pvz_errors.ErrSelectReceptionsFailed
		}
		return nil, err
	}
	return receptions, nil
}
