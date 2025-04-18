package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context,
		pvzID, productID uuid.UUID,
		dateTime time.Time,
		productType string) (uuid.UUID, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
	GetProductsByReceptionIDs(ctx context.Context, receptionIDs []*uuid.UUID) ([]oapi.Product, error)
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(dbConn *pgxpool.Pool) ProductRepository {
	return &productRepository{db: dbConn}
}

func (r *productRepository) InsertProduct(
	ctx context.Context,
	pvzID, productID uuid.UUID,
	dateTime time.Time,
	productType string,
) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var receptionID uuid.UUID
	err = tx.QueryRow(ctx, QueryInsertProduct, pvzID, productID, dateTime, productType).Scan(&receptionID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, pvz_errors.ErrNoOpenRecetionOrPvz
		}
		if pgErr.Code == "23514" {
			return uuid.Nil, pvz_errors.ErrInvalidProduct
		}
		return uuid.Nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return receptionID, nil
}

func (r *productRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	cmdTag, err := tx.Exec(ctx, QueryDeleteLastProduct, pvzID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		return pvz_errors.ErrDeletingProduct
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *productRepository) GetProductsByReceptionIDs(ctx context.Context,
	receptionIDs []*uuid.UUID) ([]oapi.Product, error) {
	rows, err := r.db.Query(ctx, QueryGetProductsByReceptions, receptionIDs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pvz_errors.ErrSelectProductsFailed
		}
		return nil, err
	}
	defer rows.Close()

	var products []oapi.Product
	for rows.Next() {
		var id uuid.UUID
		var receptionId uuid.UUID
		var dt time.Time
		var typ string
		if err := rows.Scan(&id, &receptionId, &dt, &typ); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, pvz_errors.ErrSelectProductsFailed
			}
			return nil, err
		}
		products = append(products, oapi.Product{
			Id:          &id,
			ReceptionId: receptionId,
			DateTime:    &dt,
			Type:        oapi.ProductType(typ),
		})
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pvz_errors.ErrSelectProductsFailed
		}
		return nil, err
	}
	return products, nil
}
