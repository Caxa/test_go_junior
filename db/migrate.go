package db

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ApplyMigrations() error {
	// Получаем абсолютный путь до каталога с миграциями
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationsPath := filepath.Join(dir, "migrations")

	cmd := exec.Command("migrate",
		"-path", migrationsPath,
		"-database", fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME")),
		"up")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
