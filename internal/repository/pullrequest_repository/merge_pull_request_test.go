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

func TestPullRequestStorage_MergePullRequest(t *testing.T) {
	db := setupPRTestDB(t)
	defer db.Close()
	defer cleanupPRTestDB(t, db)

	storage := NewPullRequestStorage(db)
	ctx := context.Background()

	createTestTeamAndUser(t, db, "test-team", "author1", "Author 1")

	t.Run("successful merge", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-merge",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		require.NoError(t, err)

		err = storage.MergePullRequest(ctx, "pr-merge")
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-merge")
		require.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, retrieved.Status)
		assert.NotNil(t, retrieved.MergedAt)
	})

	t.Run("idempotent merge", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:   "pr-idempotent",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)
		require.NoError(t, err)

		err = storage.MergePullRequest(ctx, "pr-idempotent")
		require.NoError(t, err)

		firstMergedAt, err := storage.GetPullRequestByID(ctx, "pr-idempotent")
		require.NoError(t, err)
		firstMergedTime := firstMergedAt.MergedAt

		err = storage.MergePullRequest(ctx, "pr-idempotent")
		require.NoError(t, err)

		retrieved, err := storage.GetPullRequestByID(ctx, "pr-idempotent")
		require.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, retrieved.Status)
		assert.Equal(t, firstMergedTime.Unix(), retrieved.MergedAt.Unix())
	})

	t.Run("merge non-existent PR", func(t *testing.T) {
		err := storage.MergePullRequest(ctx, "nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
