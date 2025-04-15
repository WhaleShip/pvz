package database

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connectPostgres(cfg Config) (*pgxpool.Pool, error) {
	maxConn := 5
	if !fiber.IsChild() {
		maxConn = runtime.NumCPU() * 4
	}

	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s pool_max_conns=%d",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
			cfg.DBName,
			maxConn,
		))
	if err != nil {
		panic(err)
	}

	if err != nil {
		return nil, fmt.Errorf("db connection failed: %w", err)
	}

	log.Println("db connected")
	return pool, nil
}
