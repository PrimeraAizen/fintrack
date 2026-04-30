package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // for migrations
	"github.com/pressly/goose/v3"
)

// RunMigrations opens the DB at dsn, sets up goose, and applies all “up” migrations.
func RunMigrations(dsn, path, tableName string) error {
	sqlDB, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	// Solves issue with CI/CD not being able to find "goose_db_version" table
	goose.SetTableName(tableName)

	if err := goose.Up(sqlDB, path); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
