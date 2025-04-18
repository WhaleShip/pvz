package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/database"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
)

type userRepository struct {
	db database.PgxIface
}

func NewUserRepository(dbConn database.PgxIface) *userRepository {
	return &userRepository{db: dbConn}
}

func (r *userRepository) InsertUser(ctx context.Context, id uuid.UUID, email, password, role string) error {
	ct, err := r.db.Exec(ctx, QueryInsertUser, id, email, password, role)
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
	err := r.db.QueryRow(ctx, QueryUserByEmail, email).Scan(&id, &password, &role)
	if err != nil {
		if errors.Is(err, r.db.ErrNoRows()) {
			return uuid.Nil, "", "", pvz_errors.ErrUserNotFound
		}
		return uuid.Nil, "", "", err
	}
	return id, password, role, nil
}
