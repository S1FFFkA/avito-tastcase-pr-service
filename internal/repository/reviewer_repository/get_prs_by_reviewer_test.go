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

func TestPrReviewersStorage_GetPRsByReviewer(t *testing.T) {
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
	})
	require.NoError(t, err)

	t.Run("no PRs found", func(t *testing.T) {
		prs, err := storage.GetPRsByReviewer(ctx, "reviewer1")
		require.NoError(t, err)
		assert.Empty(t, prs)
	})

	t.Run("successful get", func(t *testing.T) {
		pr1 := &domain.PullRequest{
			PullRequestID:   "pr1",
			PullRequestName: "PR 1",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}
		pr2 := &domain.PullRequest{
			PullRequestID:   "pr2",
			PullRequestName: "PR 2",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := prStorage.CreatePullRequestWithReviewers(ctx, pr1, []string{"reviewer1"}, false)
		require.NoError(t, err)
		err = prStorage.CreatePullRequestWithReviewers(ctx, pr2, []string{"reviewer1"}, false)
		require.NoError(t, err)

		prs, err := storage.GetPRsByReviewer(ctx, "reviewer1")
		require.NoError(t, err)
		assert.Len(t, prs, 2)
	})

	t.Run("get PRs with mixed statuses", func(t *testing.T) {
		pr1 := &domain.PullRequest{
			PullRequestID:   "pr-open",
			PullRequestName: "Open PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}
		pr2 := &domain.PullRequest{
			PullRequestID:   "pr-merged",
			PullRequestName: "Merged PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}

		err := prStorage.CreatePullRequestWithReviewers(ctx, pr1, []string{"reviewer1"}, false)
		require.NoError(t, err)
		err = prStorage.CreatePullRequestWithReviewers(ctx, pr2, []string{"reviewer1"}, false)
		require.NoError(t, err)

		err = prStorage.MergePullRequest(ctx, "pr-merged")
		require.NoError(t, err)

		prs, err := storage.GetPRsByReviewer(ctx, "reviewer1")
		require.NoError(t, err)
		assert.Len(t, prs, 2)
	})
}
