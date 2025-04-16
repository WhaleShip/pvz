package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
)

type PVZRepository interface {
	InsertPVZ(ctx context.Context, city gen.PVZCity, registrationDate time.Time) (gen.PVZ, error)
	SelectPVZ(ctx context.Context, params gen.GetPvzParams) ([]gen.PVZ, error)
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

func (r *pvzRepository) SelectPVZ(ctx context.Context, params gen.GetPvzParams) ([]gen.PVZ, error) {
	query, args := r.buildPVZQuery(params)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pvz_errors.ErrSelectPVZFailed
		}
		return nil, err
	}
	defer rows.Close()

	var pvzList []gen.PVZ
	for rows.Next() {
		var id uuid.UUID
		var city string
		var regDate time.Time
		if err = rows.Scan(&id, &city, &regDate); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, pvz_errors.ErrSelectPVZFailed
			}
			return nil, err
		}
		pvzList = append(pvzList, gen.PVZ{
			Id:               &id,
			City:             gen.PVZCity(city),
			RegistrationDate: &regDate,
		})
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pvz_errors.ErrSelectPVZFailed
		}
		return nil, err
	}
	return pvzList, nil
}

func (r *pvzRepository) buildPVZQuery(params gen.GetPvzParams) (string, []interface{}) {
	baseQuery := QuerySelectPVZ
	var args []interface{}
	var conditions []string

	if params.StartDate != nil {
		conditions = append(conditions, "registration_date >= $1")
		args = append(args, *params.StartDate)
	}
	if params.EndDate != nil {
		conditions = append(conditions, "registration_date <= $2")
		args = append(args, *params.EndDate)
	}
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	page := 1
	limit := 10
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	offset := (page - 1) * limit

	nextPlaceholder := len(args) + 1
	query += fmt.Sprintf(" ORDER BY registration_date DESC LIMIT $%d OFFSET $%d", nextPlaceholder, nextPlaceholder+1)
	args = append(args, limit, offset)

	return query, args
}
