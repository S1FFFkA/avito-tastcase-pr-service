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

func TestUserRepository_SetUserIsActive(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	t.Run("user not found", func(t *testing.T) {
		err := repo.SetUserIsActive(ctx, "nonexistent", true)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("successful activation", func(t *testing.T) {
		_, err := teamStorage.CreateTeamWithMembers(ctx, "test-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: false},
		})
		require.NoError(t, err)

		err = repo.SetUserIsActive(ctx, "user1", true)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "user1")
		require.NoError(t, err)
		assert.True(t, user.IsActive)
	})

	t.Run("successful deactivation", func(t *testing.T) {
		_, err := teamStorage.CreateTeamWithMembers(ctx, "test-team-2", []domain.TeamMember{
			{UserID: "user2", Username: "User 2", IsActive: true},
		})
		require.NoError(t, err)

		err = repo.SetUserIsActive(ctx, "user2", false)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "user2")
		require.NoError(t, err)
		assert.False(t, user.IsActive)
	})

	t.Run("idempotent activation", func(t *testing.T) {
		_, err := teamStorage.CreateTeamWithMembers(ctx, "test-team-3", []domain.TeamMember{
			{UserID: "user3", Username: "User 3", IsActive: true},
		})
		require.NoError(t, err)

		err = repo.SetUserIsActive(ctx, "user3", true)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "user3")
		require.NoError(t, err)
		assert.True(t, user.IsActive)
	})
}
