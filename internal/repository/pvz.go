package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PVZRepository interface {
}

type pvzRepository struct {
	db *pgxpool.Pool
}

func NewPVZRepository(dbConn *pgxpool.Pool) PVZRepository {
	return &pvzRepository{db: dbConn}
}
