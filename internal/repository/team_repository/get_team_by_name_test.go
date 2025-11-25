//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamStorage_GetTeamByName(t *testing.T) {
	db := setupTeamTestDB(t)
	defer db.Close()
	defer cleanupTeamTestDB(t, db)

	storage := NewTeamStorage(db)
	ctx := context.Background()

	t.Run("team not found", func(t *testing.T) {
		team, err := storage.GetTeamByName(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("successful get", func(t *testing.T) {
		teamID, err := storage.CreateTeamWithMembers(ctx, "test-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
			{UserID: "user2", Username: "User 2", IsActive: false},
		})
		require.NoError(t, err)
		require.NotEqual(t, teamID, uuid.Nil)

		team, err := storage.GetTeamByName(ctx, "test-team")
		require.NoError(t, err)
		assert.Equal(t, "test-team", team.TeamName)
		assert.Len(t, team.Members, 2)
		assert.Equal(t, "user1", team.Members[0].UserID)
		assert.Equal(t, "User 1", team.Members[0].Username)
		assert.True(t, team.Members[0].IsActive)
		assert.Equal(t, "user2", team.Members[1].UserID)
		assert.False(t, team.Members[1].IsActive)
	})

	t.Run("team with no members", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO teams (id, team_name) VALUES (gen_random_uuid(), $1)", "empty-team")
		require.NoError(t, err)

		team, err := storage.GetTeamByName(ctx, "empty-team")
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
