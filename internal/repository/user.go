package repository

import "github.com/jackc/pgx/v5/pgxpool"

type UserRepository interface {
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbConn *pgxpool.Pool) UserRepository {
	return &userRepository{db: dbConn}
}
