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

func TestPullRequestStorage_CreatePullRequestWithReviewers(t *testing.T) {
	db := setupPRTestDB(t)
	defer db.Close()
	defer cleanupPRTestDB(t, db)

	storage := NewPullRequestStorage(db)
	ctx := context.Background()

	createTestTeamAndUser(t, db, "test-team", "author1", "Author 1")
	createTestTeamAndUser(t, db, "test-team", "reviewer1", "Reviewer 1")

	t.Run("successful create", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr1",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{"reviewer1"}, false)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, "pr1", retrieved.PullRequestID)
	})

	t.Run("duplicate PR ID", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-duplicate",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		require.NoError(t, err)

		err = storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPRExists)
	})

	t.Run("create PR with need_more_reviewers", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-need-more",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{"reviewer1"}, true)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-need-more")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.NeedMoreReviewers)
		assert.True(t, *retrieved.NeedMoreReviewers)
	})
}
