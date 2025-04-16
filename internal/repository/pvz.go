package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
)

type PVZRepository interface {
	InsertPVZ(ctx context.Context, city gen.PVZCity, registrationDate time.Time) (gen.PVZ, error)
}

type pvzRepository struct {
	db *pgxpool.Pool
}

func NewPVZRepository(dbConn *pgxpool.Pool) PVZRepository {
	return &pvzRepository{db: dbConn}
}

func (r *pvzRepository) InsertPVZ(ctx context.Context, city gen.PVZCity, registrationDate time.Time) (gen.PVZ, error) {
	newPvzID := uuid.New()
	var id uuid.UUID
	var outCity string
	var regDate time.Time

	err := r.db.QueryRow(ctx, QueryInsertPVZ, newPvzID, string(city), registrationDate).
		Scan(&id, &outCity, &regDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.PVZ{}, pvz_errors.ErrInsertPVZFailed
		}
		return gen.PVZ{}, err
	}

	return gen.PVZ{
		Id:               &id,
		City:             gen.PVZCity(outCity),
		RegistrationDate: &regDate,
	}, nil
}
