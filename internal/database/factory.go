package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashborn3/BinTraceBench/internal/config"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(cfg *config.Config) (Database, error) {
	switch cfg.Database.Type {
	case "sqlite":
		return f.createSQLite(cfg)
	case "postgresql":
		return f.createPostgreSQL(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
}

func (f *Factory) createSQLite(cfg *config.Config) (Database, error) {
	dir := filepath.Dir(cfg.Database.SQLite.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create SQLite directory: %w", err)
	}

	db := NewSQLiteDB(cfg.Database.SQLite.Path)

	if err := db.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite: %w", err)
	}

	if err := db.CreateTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create SQLite tables: %w", err)
	}

	return db, nil
}

func (f *Factory) createPostgreSQL(cfg *config.Config) (Database, error) {
	db := NewPostgreSQLDB(
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	if err := db.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	if err := db.CreateTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create PostgreSQL tables: %w", err)
	}

	return db, nil
}
