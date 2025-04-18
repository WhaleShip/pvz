package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type PVZRepository interface {
	InsertPVZ(ctx context.Context, city oapi.PVZCity, registrationDate time.Time) (oapi.PVZ, error)
	SelectPVZByOpenReceptions(
		ctx context.Context,
		startDate, endDate time.Time,
		limit, offset int,
	) ([]oapi.PVZ, error)
}
type pvzRepository struct {
	db *pgxpool.Pool
}

func NewPVZRepository(dbConn *pgxpool.Pool) PVZRepository {
	return &pvzRepository{db: dbConn}
}

func (r *pvzRepository) InsertPVZ(ctx context.Context, city oapi.PVZCity, registrationDate time.Time) (oapi.PVZ, error) {
	newPvzID := uuid.New()
	var id uuid.UUID
	var outCity string
	var regDate time.Time

	err := r.db.QueryRow(ctx, QueryInsertPVZ, newPvzID, string(city), registrationDate).
		Scan(&id, &outCity, &regDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return oapi.PVZ{}, pvz_errors.ErrInsertPVZFailed
		}
		return oapi.PVZ{}, err
	}

	return oapi.PVZ{
		Id:               &id,
		City:             oapi.PVZCity(outCity),
		RegistrationDate: &regDate,
	}, nil
}

func (r *pvzRepository) SelectPVZByOpenReceptions(
	ctx context.Context,
	startDate, endDate time.Time,
	limit, offset int,
) ([]oapi.PVZ, error) {
	rows, err := r.db.Query(ctx,
		QuerySelectPvzByOpenReceptions,
		startDate, endDate,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", pvz_errors.ErrSelectPVZFailed, err)
	}
	defer rows.Close()

	var list []oapi.PVZ
	for rows.Next() {
		var id uuid.UUID
		var city string
		var regDate time.Time
		if err := rows.Scan(&id, &city, &regDate); err != nil {
			return nil, err
		}
		list = append(list, oapi.PVZ{
			Id:               &id,
			City:             oapi.PVZCity(city),
			RegistrationDate: &regDate,
		})
	}
	return list, rows.Err()
}
