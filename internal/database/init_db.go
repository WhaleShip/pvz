package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func getEnvVariable(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		log.Fatalf("enviroment error: %s variable not exist", name)
	}
	return value
}

func GetInitializedDB() (*pgxpool.Pool, error) {
	cfg := Config{
		Host:     getEnvVariable("DB_HOST"),
		Port:     getEnvVariable("DB_PORT"),
		Username: getEnvVariable("DB_USER"),
		Password: getEnvVariable("DB_PASSWORD"),
		DBName:   getEnvVariable("DB_NAME"),
		SSLMode:  getEnvVariable("SSL_MODE"),
	}
	pool, err := connectPostgres(cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("db pinged successfuly")
	return pool, nil
}
