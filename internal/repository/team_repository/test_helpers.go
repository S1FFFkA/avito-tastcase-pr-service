//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/infrastructure/database"

	"github.com/stretchr/testify/require"
)

func setupTeamTestDB(t *testing.T) *sql.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewDB(ctx)
	require.NoError(t, err)

	cleanupTeamTestDB(t, db)
	return db
}

func cleanupTeamTestDB(t *testing.T, db *sql.DB) {
	queries := []string{
		"DELETE FROM reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM users",
		"DELETE FROM teams",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		require.NoError(t, err)
	}
}
