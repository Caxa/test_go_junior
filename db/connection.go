package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	DB   *sql.DB
	once sync.Once
)

func Init() error {
	var initErr error
	once.Do(func() {
		requiredEnv := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
		for _, env := range requiredEnv {
			if os.Getenv(env) == "" {
				initErr = fmt.Errorf("missing required environment variable: %s", env)
				return
			}
		}

		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"))

		DB, initErr = sql.Open("postgres", dsn)
		if initErr != nil {
			initErr = fmt.Errorf("failed to open database connection: %w", initErr)
			return
		}

		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(5)
		DB.SetConnMaxLifetime(5 * time.Minute)
		DB.SetConnMaxIdleTime(2 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if initErr = DB.PingContext(ctx); initErr != nil {
			initErr = fmt.Errorf("database ping failed: %w", initErr)
			_ = DB.Close()
			DB = nil
		}
	})

	return initErr
}

func GetDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return DB, nil
}
