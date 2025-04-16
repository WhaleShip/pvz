package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
)

type UserRepository interface {
	InsertUser(ctx context.Context, id uuid.UUID, email, password, role string) error
	GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, string, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbConn *pgxpool.Pool) UserRepository {
	return &userRepository{db: dbConn}
}

func (r *userRepository) InsertUser(ctx context.Context, id uuid.UUID, email, password, role string) error {
	ct, err := r.db.Exec(ctx, `INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)
                       			ON CONFLICT (email) DO NOTHING`, id, email, password, role)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pvz_errors.ErrUserAlreadyExists
	}
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, string, error) {
	var id uuid.UUID
	var password string
	var role string
	err := r.db.QueryRow(ctx, `SELECT id, password, role FROM users WHERE email = $1`, email).Scan(&id, &password, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", "", pvz_errors.ErrUserNotFound
		}
		return uuid.Nil, "", "", err
	}
	return id, password, role, nil
}
