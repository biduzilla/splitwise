package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(migrations embed.FS, dsn, tableName string, logger *slog.Logger) error {
	source, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("failed to load embedded migrations: %w", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open migration database: %w", err)
	}

	instance, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: tableName,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", instance)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Info("database migrations applied successfully")
	return nil
}
