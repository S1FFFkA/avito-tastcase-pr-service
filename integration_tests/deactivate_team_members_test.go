//go:build integration

package integration_tests

import (
	"context"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	prreviewerspkg "AVITOSAMPISHU/internal/repository/reviewer_repository"
	teampkg "AVITOSAMPISHU/internal/repository/team_repository"
	repositorypkg "AVITOSAMPISHU/internal/repository/user_repository"
	userservice "AVITOSAMPISHU/internal/service/user_service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegrationDeactivateTeamMembersFlow(t *testing.T) {
	truncateAll(t)

	ctx := context.Background()

	teamID := uuid.New()
	teamName := "team-alpha"

	_, err := testDB.ExecContext(ctx, `INSERT INTO teams (id, team_name) VALUES ($1, $2)`, teamID, teamName)
	require.NoError(t, err)

	users := []struct {
		ID     string
		Name   string
		Active bool
	}{
		{"u1", "Alice", true},
		{"u2", "Bob", true},
		{"u3", "Charlie", true},
		{"u4", "Diana", true},
	}

	for _, u := range users {
		_, err = testDB.ExecContext(ctx,
			`INSERT INTO users (id, username, team_id, is_active) VALUES ($1, $2, $3, $4)`,
			u.ID, u.Name, teamID, u.Active,
		)
		require.NoError(t, err)
	}

	prID := "pr-200"
	_, err = testDB.ExecContext(ctx,
		`INSERT INTO pull_requests (id, pull_requests_name, author_id, status) VALUES ($1, $2, $3, 'OPEN')`,
		prID, "Refactor module", "u1",
	)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		`INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2), ($1, $3)`,
		prID, "u2", "u3",
	)
	require.NoError(t, err)

	userRepo := repositorypkg.NewUserRepository(testDB)
	prRepo := prreviewerspkg.NewPrReviewersStorage(testDB)
	teamRepo := teampkg.NewTeamStorage(testDB)
	userService := userservice.NewUserService(userRepo, prRepo, teamRepo)

	res, err := userService.DeactivateTeamMembers(ctx, &domain.DeactivateTeamMembersReq{
		TeamName: teamName,
		UserIDs:  []string{"u2"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"u2"}, res.DeactivatedUserIDs)
	require.Len(t, res.Reassignments, 1)
	require.Equal(t, prID, res.Reassignments[0].PrID)
	require.Equal(t, "u2", res.Reassignments[0].OldReviewerID)
	// NewReviewerID might be empty if no suitable candidate is found, or a new one
	// We only assert it's not the old reviewer and is one of the remaining active users
	if res.Reassignments[0].NewReviewerID != "" {
		require.NotEqual(t, "u2", res.Reassignments[0].NewReviewerID)
		require.Contains(t, []string{"u3", "u4"}, res.Reassignments[0].NewReviewerID)
	}

	var isActive bool
	err = testDB.QueryRowContext(ctx, `SELECT is_active FROM users WHERE id = $1`, "u2").Scan(&isActive)
	require.NoError(t, err)
	require.False(t, isActive, "user must be deactivated in DB")

	var remainingReviewers []string
	rows, err := testDB.QueryContext(ctx, `SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1`, prID)
	require.NoError(t, err)
	for rows.Next() {
		var reviewerID string
		require.NoError(t, rows.Scan(&reviewerID))
		remainingReviewers = append(remainingReviewers, reviewerID)
	}
	require.NoError(t, rows.Err())
	require.NotContains(t, remainingReviewers, "u2", "u2 should be removed from reviewers")
	require.Contains(t, remainingReviewers, "u3", "u3 should remain as reviewer")
	// Если был назначен новый ревьювер, проверяем что он в списке
	if res.Reassignments[0].NewReviewerID != "" {
		require.Contains(t, remainingReviewers, res.Reassignments[0].NewReviewerID, "new reviewer should be in the list")
		require.Len(t, remainingReviewers, 2, "should have u3 and new reviewer")
	} else {
		// Если новый ревьювер не был назначен, остается только u3
		require.Len(t, remainingReviewers, 1, "should have only u3 if no new reviewer assigned")
	}
}

func TestIntegrationDeactivateTeamMembers_Rollback(t *testing.T) {
	truncateAll(t)

	ctx := context.Background()
	teamID := uuid.New()
	teamName := "team-rollback"

	_, err := testDB.ExecContext(ctx, `INSERT INTO teams (id, team_name) VALUES ($1, $2)`, teamID, teamName)
	require.NoError(t, err)

	members := []struct {
		ID   string
		Name string
	}{
		{"u1", "Alice"},
		{"u2", "Bob"},
	}

	for _, m := range members {
		_, err = testDB.ExecContext(ctx,
			`INSERT INTO users (id, username, team_id, is_active) VALUES ($1, $2, $3, true)`,
			m.ID, m.Name, teamID,
		)
		require.NoError(t, err)
	}

	prID := "pr-rollback"
	_, err = testDB.ExecContext(ctx,
		`INSERT INTO pull_requests (id, pull_requests_name, author_id, status) VALUES ($1, $2, $3, 'OPEN')`,
		prID, "Rollback check", "u1",
	)
	require.NoError(t, err)

	_, err = testDB.ExecContext(ctx,
		`INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`,
		prID, "u2",
	)
	require.NoError(t, err)

	// Simulate a failed reassignment by providing a non-existent new reviewer ID
	// This will cause a foreign key violation in the repository layer, triggering a rollback
	teamRepo := teampkg.NewTeamStorage(testDB)
	reassignments := []domain.ReviewerReassignment{
		{PrID: prID, OldReviewerID: "u2", NewReviewerID: "non-existent-user"},
	}

	// Call the repository method directly to trigger the rollback logic
	deactivatedIDs, err := teamRepo.DeactivateTeamMembers(ctx, teamName, []string{"u2"}, reassignments)
	require.Error(t, err)
	require.Contains(t, err.Error(), "foreign key constraint")
	require.Empty(t, deactivatedIDs)

	// Verify that u2 is still active (rollback occurred)
	var isActive bool
	err = testDB.QueryRowContext(ctx, `SELECT is_active FROM users WHERE id = $1`, "u2").Scan(&isActive)
	require.NoError(t, err)
	require.True(t, isActive, "user should remain active due to rollback")

	// Verify that the reviewer 'u2' is still assigned to pr-rollback (rollback occurred)
	var reviewerCount int
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`, prID, "u2").Scan(&reviewerCount)
	require.NoError(t, err)
	require.Equal(t, 1, reviewerCount, "reviewer 'u2' should still be assigned due to rollback")
}

func TestIntegrationDeactivateTeamMembers_CannotDeactivateAll(t *testing.T) {
	truncateAll(t)

	ctx := context.Background()

	teamID := uuid.New()
	teamName := "team-no-deactivate-all"

	_, err := testDB.ExecContext(ctx, `INSERT INTO teams (id, team_name) VALUES ($1, $2)`, teamID, teamName)
	require.NoError(t, err)

	users := []struct {
		ID     string
		Name   string
		Active bool
	}{
		{"u1", "Alice", true},
		{"u2", "Bob", true},
	}

	for _, u := range users {
		_, err = testDB.ExecContext(ctx,
			`INSERT INTO users (id, username, team_id, is_active) VALUES ($1, $2, $3, $4)`,
			u.ID, u.Name, teamID, u.Active,
		)
		require.NoError(t, err)
	}

	userRepo := repositorypkg.NewUserRepository(testDB)
	prRepo := prreviewerspkg.NewPrReviewersStorage(testDB)
	teamRepo := teampkg.NewTeamStorage(testDB)
	userService := userservice.NewUserService(userRepo, prRepo, teamRepo)

	// Test case 1: Empty UserIDs list
	_, err = userService.DeactivateTeamMembers(ctx, &domain.DeactivateTeamMembersReq{
		TeamName: teamName,
		UserIDs:  []string{}, // Empty list
	})
	require.ErrorIs(t, err, domain.ErrInvalidRequest)
	require.Contains(t, err.Error(), "cannot deactivate all team members without explicit user IDs")

	// Test case 2: All team members in UserIDs list
	_, err = userService.DeactivateTeamMembers(ctx, &domain.DeactivateTeamMembersReq{
		TeamName: teamName,
		UserIDs:  []string{"u1", "u2"}, // All members
	})
	require.ErrorIs(t, err, domain.ErrInvalidRequest)
	require.Contains(t, err.Error(), "cannot deactivate all team members, team would be left without reviewers")
}
