//go:build integration

package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/infrastructure/database"
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	logger.InitLogger()
	logger.Logger.Infow("Starting integration tests setup...")

	cfg := database.Config{
		Host:            helpers.EnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            helpers.EnvOrDefault("TEST_DB_PORT", "55432"),
		User:            helpers.EnvOrDefault("TEST_DB_USER", "test_user"),
		Password:        helpers.EnvOrDefault("TEST_DB_PASSWORD", "test_password"),
		DBName:          helpers.EnvOrDefault("TEST_DB_NAME", "avito_pr_test"),
		SSLMode:         helpers.EnvOrDefault("TEST_DB_SSLMODE", "disable"),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	var err error
	testDB, err = database.NewDBWithConfig(context.Background(), cfg)
	if err != nil {
		logger.Logger.Fatalw("Failed to connect to test database", "error", err)
	}
	defer testDB.Close()

	logger.Logger.Infow("Test database connection established")

	// Run tests
	exitCode := m.Run()

	logger.Logger.Infow("Integration tests finished. Cleaning up...")
	os.Exit(exitCode)
}

func truncateAll(t *testing.T) {
	tables := make([]string, 0, 4)
	tables = append(tables, "reviewers", "pull_requests", "users", "teams")
	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(t, err, "Failed to truncate table %s", table)
	}
}
