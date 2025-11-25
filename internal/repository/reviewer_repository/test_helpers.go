//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/infrastructure/database"
	pullrequest_repository "AVITOSAMPISHU/internal/repository/pullrequest_repository"

	"github.com/stretchr/testify/require"
)

func setupReviewerTestDB(t *testing.T) *sql.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewDB(ctx)
	require.NoError(t, err)

	cleanupReviewerTestDB(t, db)
	return db
}

func cleanupReviewerTestDB(t *testing.T, db *sql.DB) {
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

func createTestPR(t *testing.T, db *sql.DB, prID, authorID string) {
	prStorage := pullrequest_repository.NewPullRequestStorage(db)
	ctx := context.Background()
	pr := &domain.PullRequest{
		PullRequestID:   prID,
		PullRequestName: "Test PR",
		AuthorID:        authorID,
		Status:          domain.PRStatusOpen,
	}
	err := prStorage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
	require.NoError(t, err)
}
