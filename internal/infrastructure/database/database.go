package database

import (
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (c Config) buildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func NewDB(ctx context.Context) (*sql.DB, error) {
	config := Config{
		Host:            helpers.EnvOrDefault("DB_HOST", "localhost"),
		Port:            helpers.EnvOrDefault("DB_PORT", "5432"),
		User:            helpers.EnvOrDefault("DB_USER", "avito_user"),
		Password:        helpers.EnvOrDefault("DB_PASSWORD", "avito_password"),
		DBName:          helpers.EnvOrDefault("DB_NAME", "avito_pr_db"),
		SSLMode:         helpers.EnvOrDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:    defaultMaxOpenConns,
		MaxIdleConns:    defaultMaxIdleConns,
		ConnMaxLifetime: defaultConnMaxLifetime,
		ConnMaxIdleTime: defaultConnMaxIdleTime,
	}

	return NewDBWithConfig(ctx, config)
}

// NewDBWithConfig создает подключение к БД с заданной конфигурацией
func NewDBWithConfig(ctx context.Context, cfg Config) (*sql.DB, error) {
	dsn := cfg.buildDSN()
	logger.Logger.Infow("connecting to database", "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName, "user", cfg.User)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Logger.Errorw("failed to open database connection", "error", err, "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logger.Logger.Infow("pinging database", "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
	if err = db.PingContext(pingCtx); err != nil {
		db.Close()
		logger.Logger.Errorw("failed to ping database", "error", err, "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Logger.Infow("database connection established", "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
	return db, nil
}
