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

func TestPullRequestStorage_GetPullRequestByID(t *testing.T) {
	db := setupPRTestDB(t)
	defer db.Close()
	defer cleanupPRTestDB(t, db)

	storage := NewPullRequestStorage(db)
	ctx := context.Background()

	createTestTeamAndUser(t, db, "test-team", "author1", "Author 1")

	t.Run("PR not found", func(t *testing.T) {
		pr, err := storage.GetPullRequestByID(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, pr)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("successful get", func(t *testing.T) {
		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:   "pr1",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, "pr1", retrieved.PullRequestID)
		assert.Equal(t, "Test PR", retrieved.PullRequestName)
		assert.Equal(t, "author1", retrieved.AuthorID)
		assert.Equal(t, domain.PRStatusOpen, retrieved.Status)
		assert.NotNil(t, retrieved.CreatedAt)
		assert.True(t, retrieved.CreatedAt.After(now.Add(-time.Second)))
	})

	t.Run("get PR with reviewers", func(t *testing.T) {
		createTestTeamAndUser(t, db, "test-team-2", "reviewer1", "Reviewer 1")
		createTestTeamAndUser(t, db, "test-team-2", "reviewer2", "Reviewer 2")

		pr := &domain.PullRequest{
			PullRequestID:   "pr-with-reviewers",
			PullRequestName: "PR With Reviewers",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{"reviewer1", "reviewer2"}, false)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-with-reviewers")
		require.NoError(t, err)
		assert.Len(t, retrieved.AssignedReviewers, 2)
		assert.Contains(t, retrieved.AssignedReviewers, "reviewer1")
		assert.Contains(t, retrieved.AssignedReviewers, "reviewer2")
	})
}
