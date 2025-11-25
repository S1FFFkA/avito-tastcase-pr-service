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

func setupUserTestDB(t *testing.T) *sql.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewDB(ctx)
	require.NoError(t, err)

	cleanupUserTestDB(t, db)
	return db
}

func cleanupUserTestDB(t *testing.T, db *sql.DB) {
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

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)
	teamStorage := NewTeamStorage(db)
	ctx := context.Background()

	t.Run("user not found", func(t *testing.T) {
		user, err := repo.GetUserByID(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("successful get", func(t *testing.T) {
		_, err := teamStorage.CreateTeamWithMembers(ctx, "test-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
		})
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, "user1", user.UserID)
		assert.Equal(t, "User 1", user.Username)
		assert.Equal(t, "test-team", user.TeamName)
		assert.True(t, user.IsActive)
	})

	t.Run("user without team", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO users (id, username, is_active) VALUES ($1, $2, $3)", "user-no-team", "User No Team", true)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "user-no-team")
		require.NoError(t, err)
		assert.Equal(t, "user-no-team", user.UserID)
		assert.Equal(t, "", user.TeamName)
	})
}
