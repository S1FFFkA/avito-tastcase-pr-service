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

func TestTeamStorage_CreateTeamWithMembers(t *testing.T) {
	db := setupTeamTestDB(t)
	defer db.Close()
	defer cleanupTeamTestDB(t, db)

	storage := NewTeamStorage(db)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		teamID, err := storage.CreateTeamWithMembers(ctx, "new-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, teamID)

		team, err := storage.GetTeamByName(ctx, "new-team")
		require.NoError(t, err)
		assert.Equal(t, "new-team", team.TeamName)
		assert.Len(t, team.Members, 1)
	})

	t.Run("duplicate team name", func(t *testing.T) {
		_, err := storage.CreateTeamWithMembers(ctx, "duplicate-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
		})
		require.NoError(t, err)

		_, err = storage.CreateTeamWithMembers(ctx, "duplicate-team", []domain.TeamMember{
			{UserID: "user2", Username: "User 2", IsActive: true},
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTeamExists)
	})

	t.Run("create team with multiple members", func(t *testing.T) {
		teamID, err := storage.CreateTeamWithMembers(ctx, "multi-member-team", []domain.TeamMember{
			{UserID: "user1", Username: "User 1", IsActive: true},
			{UserID: "user2", Username: "User 2", IsActive: true},
			{UserID: "user3", Username: "User 3", IsActive: false},
		})
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, teamID)

		team, err := storage.GetTeamByName(ctx, "multi-member-team")
		require.NoError(t, err)
		assert.Len(t, team.Members, 3)
	})
}
