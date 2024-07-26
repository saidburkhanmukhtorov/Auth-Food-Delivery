package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
)

// Connect establishes a connection to the PostgreSQL database.
func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// // Test the connection
	// if err := pool.Ping(context.Background()); err != nil {
	// 	return nil, fmt.Errorf("database ping failed: %w", err)
	// }

	log.Println("Connected to PostgreSQL database")
	return pool, nil
}
