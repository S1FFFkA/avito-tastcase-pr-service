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

func TestPullRequestStorage_SetNeedMoreReviewers(t *testing.T) {
	db := setupPRTestDB(t)
	defer db.Close()
	defer cleanupPRTestDB(t, db)

	storage := NewPullRequestStorage(db)
	ctx := context.Background()

	createTestTeamAndUser(t, db, "test-team", "author1", "Author 1")

	t.Run("successful update", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-need-more",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		require.NoError(t, err)

		err = storage.SetNeedMoreReviewers(ctx, "pr-need-more", true)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-need-more")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.NeedMoreReviewers)
		assert.True(t, *retrieved.NeedMoreReviewers)
	})

	t.Run("PR not found", func(t *testing.T) {
		err := storage.SetNeedMoreReviewers(ctx, "nonexistent", true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("update need_more_reviewers to false", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-need-more-false",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, true)
		require.NoError(t, err)

		err = storage.SetNeedMoreReviewers(ctx, "pr-need-more-false", false)
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-need-more-false")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.NeedMoreReviewers)
		assert.False(t, *retrieved.NeedMoreReviewers)
	})
}
