//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/infrastructure/database"
	team_repository "AVITOSAMPISHU/internal/repository/team_repository"

	"github.com/stretchr/testify/require"
)

func setupPRTestDB(t *testing.T) *sql.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewDB(ctx)
	require.NoError(t, err)

	cleanupPRTestDB(t, db)
	return db
}

func cleanupPRTestDB(t *testing.T, db *sql.DB) {
	queries := make([]string, 0, 4)
	queries = append(queries,
		"DELETE FROM reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM users",
		"DELETE FROM teams",
	)

	for _, query := range queries {
		_, err := db.Exec(query)
		require.NoError(t, err)
	}
}

func createTestTeamAndUser(t *testing.T, db *sql.DB, teamName, userID, username string) {
	teamStorage := team_repository.NewTeamStorage(db)
	ctx := context.Background()
	_, err := teamStorage.CreateTeamWithMembers(ctx, teamName, []domain.TeamMember{
		{UserID: userID, Username: username, IsActive: true},
	})
	require.NoError(t, err)
}
