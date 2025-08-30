package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type DatabaseConfig struct {
	Type     string         `json:"type"` // "sqlite" or "postgresql"
	SQLite   SQLiteConfig   `json:"sqlite"`
	Postgres PostgresConfig `json:"postgres"`
}

type SQLiteConfig struct {
	Path string `json:"path"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type AuthConfig struct {
	SessionExpiry int `json:"session_expiry"` // in hours
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "sqlite"),
			SQLite: SQLiteConfig{
				Path: getEnv("SQLITE_PATH", "./data/bintracebench.db"),
			},
			Postgres: PostgresConfig{
				Host:     getEnv("POSTGRES_HOST", "localhost"),
				Port:     getEnvAsInt("POSTGRES_PORT", 5432),
				User:     getEnv("POSTGRES_USER", "bintracebench"),
				Password: getEnv("POSTGRES_PASSWORD", ""),
				DBName:   getEnv("POSTGRES_DB", "bintracebench"),
				SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			},
		},
		Auth: AuthConfig{
			SessionExpiry: getEnvAsInt("SESSION_EXPIRY_HOURS", 24),
		},
	}
}

func (c *Config) Validate() error {
	if c.Database.Type != "sqlite" && c.Database.Type != "postgresql" {
		return fmt.Errorf("invalid database type: %s (must be 'sqlite' or 'postgresql')", c.Database.Type)
	}

	if c.Database.Type == "postgresql" {
		if c.Database.Postgres.Host == "" {
			return fmt.Errorf("PostgreSQL host is required")
		}
		if c.Database.Postgres.User == "" {
			return fmt.Errorf("PostgreSQL user is required")
		}
		if c.Database.Postgres.DBName == "" {
			return fmt.Errorf("PostgreSQL database name is required")
		}
	}

	if c.Database.Type == "sqlite" {
		if c.Database.SQLite.Path == "" {
			return fmt.Errorf("SQLite path is required")
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
