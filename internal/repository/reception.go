package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceptionRepository interface {
}

type receptionRepository struct {
	db *pgxpool.Pool
}

func NewReceptionRepository(dbConn *pgxpool.Pool) ReceptionRepository {
	return &receptionRepository{db: dbConn}
}
