//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/infrastructure/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrReviewersStorage_GetAssignedReviewers(t *testing.T) {
	db := setupReviewerTestDB(t)
	defer db.Close()
	defer cleanupReviewerTestDB(t, db)

	storage := NewPrReviewersStorage(db)
	prStorage := NewPullRequestStorage(db)
	ctx := context.Background()

	teamStorage := NewTeamStorage(db)
	_, err := teamStorage.CreateTeamWithMembers(ctx, "test-team", []domain.TeamMember{
		{UserID: "author1", Username: "Author 1", IsActive: true},
		{UserID: "reviewer1", Username: "Reviewer 1", IsActive: true},
		{UserID: "reviewer2", Username: "Reviewer 2", IsActive: true},
	})
	require.NoError(t, err)

	t.Run("no reviewers assigned", func(t *testing.T) {
		createTestPR(t, db, "pr1", "author1")

		reviewers, err := storage.GetAssignedReviewers(ctx, "pr1")
		require.NoError(t, err)
		assert.Empty(t, reviewers)
	})

	t.Run("successful get", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr2",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := prStorage.CreatePullRequestWithReviewers(ctx, pr, []string{"reviewer1", "reviewer2"}, false)
		require.NoError(t, err)

		reviewers, err := storage.GetAssignedReviewers(ctx, "pr2")
		require.NoError(t, err)
		assert.Len(t, reviewers, 2)
		assert.Contains(t, reviewers, "reviewer1")
		assert.Contains(t, reviewers, "reviewer2")
	})

	t.Run("get reviewers for non-existent PR", func(t *testing.T) {
		reviewers, err := storage.GetAssignedReviewers(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, reviewers)
	})
}
