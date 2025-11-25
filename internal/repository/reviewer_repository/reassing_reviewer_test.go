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

func TestPrReviewersStorage_ReassignReviewer(t *testing.T) {
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

	t.Run("successful reassignment", func(t *testing.T) {
		createTestPR(t, db, "pr1", "author1")
		_, err = db.Exec("INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", "pr1", "reviewer1")
		require.NoError(t, err)

		err = storage.ReassignReviewer(ctx, "pr1", "reviewer1", "reviewer2")
		require.NoError(t, err)

		reviewers, err := storage.GetAssignedReviewers(ctx, "pr1")
		require.NoError(t, err)
		assert.Len(t, reviewers, 1)
		assert.NotContains(t, reviewers, "reviewer1")
		assert.Contains(t, reviewers, "reviewer2")
	})

	t.Run("reassign non-existent old reviewer", func(t *testing.T) {
		createTestPR(t, db, "pr2", "author1")

		err = storage.ReassignReviewer(ctx, "pr2", "non-existent-reviewer", "reviewer2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotAssigned)
	})

	t.Run("reassign on merged PR", func(t *testing.T) {
		createTestPR(t, db, "pr-merged", "author1")
		_, err = db.Exec("INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", "pr-merged", "reviewer1")
		require.NoError(t, err)

		err = prStorage.MergePullRequest(ctx, "pr-merged")
		require.NoError(t, err)

		err = storage.ReassignReviewer(ctx, "pr-merged", "reviewer1", "reviewer2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPRMerged)
	})

	t.Run("remove reviewer without replacement", func(t *testing.T) {
		createTestPR(t, db, "pr-remove", "author1")
		_, err = db.Exec("INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", "pr-remove", "reviewer1")
		require.NoError(t, err)

		err = storage.ReassignReviewer(ctx, "pr-remove", "reviewer1", "")
		require.NoError(t, err)

		reviewers, err := storage.GetAssignedReviewers(ctx, "pr-remove")
		require.NoError(t, err)
		assert.Empty(t, reviewers)
	})

	t.Run("reassign on non-existent PR", func(t *testing.T) {
		err = storage.ReassignReviewer(ctx, "nonexistent", "reviewer1", "reviewer2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
