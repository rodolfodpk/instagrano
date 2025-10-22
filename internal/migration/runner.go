package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// Runner handles database migrations
type Runner struct {
	logger *zap.Logger
}

// NewRunner creates a new migration runner
func NewRunner(logger *zap.Logger) *Runner {
	return &Runner{
		logger: logger,
	}
}

// RunMigrations runs all pending migrations
func (r *Runner) RunMigrations(dbURL, migrationsPath string) error {
	// Ensure migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	// Convert to absolute path for migrate
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create a separate database connection for migrations
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	r.logger.Info("running database migrations", zap.String("path", absPath))

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			r.logger.Info("database is up to date, no migrations to run")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	r.logger.Info("database migrations completed successfully")
	return nil
}

// CreateDatabaseIfNotExists creates the database if it doesn't exist
func (r *Runner) CreateDatabaseIfNotExists(dbURL, dbName string) error {
	// Parse the database URL to extract components
	// Replace the database name with 'postgres' to connect to the default database
	postgresURL := strings.Replace(dbURL, "/"+dbName, "/postgres", 1)

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		r.logger.Info("creating database", zap.String("database", dbName))
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		r.logger.Info("database created successfully", zap.String("database", dbName))
	} else {
		r.logger.Info("database already exists", zap.String("database", dbName))
	}

	return nil
}
