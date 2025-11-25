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

func TestTeamStorage_DeactivateTeamMembers(t *testing.T) {
	db := setupTeamTestDB(t)
	defer db.Close()
	defer cleanupTeamTestDB(t, db)

	storage := NewTeamStorage(db)
	ctx := context.Background()

	_, err := storage.CreateTeamWithMembers(ctx, "deactivate-team", []domain.TeamMember{
		{UserID: "user1", Username: "User 1", IsActive: true},
		{UserID: "user2", Username: "User 2", IsActive: true},
		{UserID: "user3", Username: "User 3", IsActive: true},
	})
	require.NoError(t, err)

	t.Run("deactivate specific users", func(t *testing.T) {
		deactivated, err := storage.DeactivateTeamMembers(ctx, "deactivate-team", []string{"user1", "user2"}, nil)
		require.NoError(t, err)
		assert.Len(t, deactivated, 2)
		assert.Contains(t, deactivated, "user1")
		assert.Contains(t, deactivated, "user2")

		team, err := storage.GetTeamByName(ctx, "deactivate-team")
		require.NoError(t, err)
		for _, member := range team.Members {
			if member.UserID == "user1" || member.UserID == "user2" {
				assert.False(t, member.IsActive)
			}
			if member.UserID == "user3" {
				assert.True(t, member.IsActive)
			}
		}
	})

	t.Run("deactivate all users in team", func(t *testing.T) {
		_, err := storage.CreateTeamWithMembers(ctx, "all-deactivate-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
			{UserID: "user2", Username: "User 2", IsActive: true},
		})
		require.NoError(t, err)

		deactivated, err := storage.DeactivateTeamMembers(ctx, "all-deactivate-team", []string{"user1", "user2"}, nil)
		require.NoError(t, err)
		assert.Len(t, deactivated, 2)
	})

	t.Run("deactivate with reassignments", func(t *testing.T) {
		_, err := storage.CreateTeamWithMembers(ctx, "reassign-team", []domain.TeamMember{
			{UserID: "reviewer1", Username: "Reviewer 1", IsActive: true},
			{UserID: "reviewer2", Username: "Reviewer 2", IsActive: true},
		})
		require.NoError(t, err)

		prStorage := NewPullRequestStorage(db)
		pr := &domain.PullRequest{
			PullRequestID:   "pr-reassign",
			PullRequestName: "Test PR",
			AuthorID:        "author1",
			Status:          domain.PRStatusOpen,
		}
		err = prStorage.CreatePullRequestWithReviewers(ctx, pr, []string{"reviewer1"}, false)
		require.NoError(t, err)

		reassignments := []domain.ReviewerReassignment{
			{PrID: "pr-reassign", OldReviewerID: "reviewer1", NewReviewerID: "reviewer2"},
		}

		deactivated, err := storage.DeactivateTeamMembers(ctx, "reassign-team", []string{"reviewer1"}, reassignments)
		require.NoError(t, err)
		assert.Len(t, deactivated, 1)
		assert.Contains(t, deactivated, "reviewer1")
	})
}
