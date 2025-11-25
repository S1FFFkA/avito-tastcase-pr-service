//go:build integration

package integration_tests

import (
	"context"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	pullrequest_repository "AVITOSAMPISHU/internal/repository/pullrequest_repository"
	reviewer_repository "AVITOSAMPISHU/internal/repository/reviewer_repository"
	team_repository "AVITOSAMPISHU/internal/repository/team_repository"
	user_repository "AVITOSAMPISHU/internal/repository/user_repository"
	pullrequest_service "AVITOSAMPISHU/internal/service/pullrequest_service"
	team_service "AVITOSAMPISHU/internal/service/team_service"

	"github.com/stretchr/testify/require"
)

func TestIntegrationPullRequestFlow(t *testing.T) {
	truncateAll(t)

	ctx := context.Background()

	// Setup Repositories
	userRepo := user_repository.NewUserRepository(testDB)
	teamRepo := team_repository.NewTeamStorage(testDB)
	prRepo := pullrequest_repository.NewPullRequestStorage(testDB)
	prReviewersRepo := reviewer_repository.NewPrReviewersStorage(testDB)

	// Setup Services
	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// 1. Create Team
	teamName := "dev-team"
	authorID := "author-1"
	reviewer1ID := "reviewer-1"
	reviewer2ID := "reviewer-2"
	reviewer3ID := "reviewer-3" // Extra reviewer for reassign
	inactiveUserID := "inactive-user-1"

	_, err := teamSvc.CreateTeam(ctx, &domain.Team{
		TeamName: teamName,
		Members: []domain.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: reviewer1ID, Username: "Reviewer1", IsActive: true},
			{UserID: reviewer2ID, Username: "Reviewer2", IsActive: true},
			{UserID: reviewer3ID, Username: "Reviewer3", IsActive: true},
			{UserID: inactiveUserID, Username: "Inactive", IsActive: false},
		},
	})
	require.NoError(t, err)

	// 2. Create Pull Request
	prID := "pr-123"
	prName := "Feature X"
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   prID,
		PullRequestName: prName,
		AuthorID:        authorID,
	}
	createdPR, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err)
	require.NotNil(t, createdPR)
	require.Equal(t, prID, createdPR.PullRequestID)
	require.Equal(t, domain.PRStatusOpen, createdPR.Status)
	require.Len(t, createdPR.AssignedReviewers, domain.MaxReviewersCount)
	require.NotContains(t, createdPR.AssignedReviewers, authorID)
	require.NotContains(t, createdPR.AssignedReviewers, inactiveUserID) // Inactive user should not be selected
	// Проверяем, что выбраны только активные ревьюверы (не автор и не неактивный)
	for _, reviewerID := range createdPR.AssignedReviewers {
		require.NotEqual(t, authorID, reviewerID, "Author should not be selected as reviewer")
		require.NotEqual(t, inactiveUserID, reviewerID, "Inactive user should not be selected as reviewer")
	}

	// Verify reviewers in DB
	dbReviewers, err := prReviewersRepo.GetAssignedReviewers(ctx, prID)
	require.NoError(t, err)
	require.Len(t, dbReviewers, domain.MaxReviewersCount)
	require.NotContains(t, dbReviewers, authorID)

	// 3. Reassign Reviewer - используем первого реально назначенного ревьювера
	require.NotEmpty(t, createdPR.AssignedReviewers, "PR should have assigned reviewers")
	oldReviewer := createdPR.AssignedReviewers[0]
	reassignReq := &domain.ReassignReviewerReq{
		PullRequestID: prID,
		OldUserID:     oldReviewer,
	}
	reassignedPR, newReviewer, err := prSvc.ReassignReviewer(ctx, reassignReq)
	require.NoError(t, err)
	require.NotNil(t, reassignedPR)
	require.NotEqual(t, oldReviewer, newReviewer)
	require.Contains(t, reassignedPR.AssignedReviewers, newReviewer)
	require.NotContains(t, reassignedPR.AssignedReviewers, oldReviewer)

	// Verify reviewers in DB after reassign
	dbReviewersAfterReassign, err := prReviewersRepo.GetAssignedReviewers(ctx, prID)
	require.NoError(t, err)
	require.Len(t, dbReviewersAfterReassign, domain.MaxReviewersCount)
	require.NotContains(t, dbReviewersAfterReassign, oldReviewer)
	require.Contains(t, dbReviewersAfterReassign, newReviewer)

	// 4. Merge Pull Request
	mergeReq := &domain.MergePullRequestReq{PullRequestID: prID}
	mergedPR, err := prSvc.MergePullRequest(ctx, mergeReq)
	require.NoError(t, err)
	require.NotNil(t, mergedPR)
	require.Equal(t, domain.PRStatusMerged, mergedPR.Status)
	require.NotNil(t, mergedPR.MergedAt)

	// Verify PR status in DB
	dbPR, err := prRepo.GetPullRequestByID(ctx, prID)
	require.NoError(t, err)
	require.Equal(t, domain.PRStatusMerged, dbPR.Status)
	require.NotNil(t, dbPR.MergedAt)

	// 5. Attempt to reassign on merged PR (negative)
	_, _, err = prSvc.ReassignReviewer(ctx, reassignReq)
	require.ErrorIs(t, err, domain.ErrPRMerged)

	// 6. Cascade Delete (delete team, check users and PRs are gone)
	// Деактивируем всех активных пользователей (4 активных, но не всех 5, чтобы не нарушить проверку)
	_, err = teamRepo.DeactivateTeamMembers(ctx, teamName, []string{authorID, reviewer1ID, reviewer2ID, reviewer3ID}, nil)
	require.NoError(t, err)

	// Verify users are deactivated (not deleted, just is_active = false)
	var activeUserCount int
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE team_id = (SELECT id FROM teams WHERE team_name = $1) AND is_active = true`, teamName).Scan(&activeUserCount)
	require.NoError(t, err)
	require.Equal(t, 0, activeUserCount, "All active users should be deactivated")

	// Verify inactive user is still there
	var inactiveUserCount int
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE team_id = (SELECT id FROM teams WHERE team_name = $1) AND is_active = false`, teamName).Scan(&inactiveUserCount)
	require.NoError(t, err)
	require.Equal(t, 5, inactiveUserCount, "All users should be deactivated (including previously inactive)")

	// Verify PR is still there (deactivation doesn't delete PRs)
	var prCount int
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pull_requests WHERE id = $1`, prID).Scan(&prCount)
	require.NoError(t, err)
	require.Equal(t, 1, prCount, "PR should still exist after user deactivation")

	// Now delete the team to trigger cascade delete
	_, err = testDB.ExecContext(ctx, `DELETE FROM teams WHERE team_name = $1`, teamName)
	require.NoError(t, err)

	// Verify users are gone (cascade delete should have removed them)
	var userCount int
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE team_id IN (SELECT id FROM teams WHERE team_name = $1)`, teamName).Scan(&userCount)
	require.NoError(t, err)
	require.Equal(t, 0, userCount, "All users should be deleted via cascade after team deletion")

	// Verify PR is gone
	err = testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pull_requests WHERE id = $1`, prID).Scan(&prCount)
	require.NoError(t, err)
	require.Equal(t, 0, prCount, "PR should be deleted via cascade after team deletion")
}
