package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(dbConn *pgxpool.Pool) ProductRepository {
	return &productRepository{db: dbConn}
}
