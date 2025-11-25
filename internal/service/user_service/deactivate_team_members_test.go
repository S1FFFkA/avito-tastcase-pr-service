package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type mockUserRepositoryForDeactivate struct {
	repository.UserRepositoryInterface
	getUserByIDFunc func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockUserRepositoryForDeactivate) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

type mockPrReviewersRepositoryForDeactivate struct {
	repository.PrReviewersRepositoryInterface
	getPRsByReviewerFunc     func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	getAssignedReviewersFunc func(ctx context.Context, prID string) ([]string, error)
}

func (m *mockPrReviewersRepositoryForDeactivate) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	if m.getAssignedReviewersFunc != nil {
		return m.getAssignedReviewersFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPrReviewersRepositoryForDeactivate) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return []domain.PullRequestShort{}, nil
}

type mockTeamRepositoryForDeactivate struct {
	repository.TeamRepositoryInterface
	getTeamByNameFunc         func(ctx context.Context, teamName string) (*domain.Team, error)
	deactivateTeamMembersFunc func(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error)
}

func (m *mockTeamRepositoryForDeactivate) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.getTeamByNameFunc != nil {
		return m.getTeamByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func (m *mockTeamRepositoryForDeactivate) DeactivateTeamMembers(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
	if m.deactivateTeamMembersFunc != nil {
		return m.deactivateTeamMembersFunc(ctx, teamName, userIDs, reassignments)
	}
	return nil, nil
}

func TestUserService_DeactivateTeamMembers(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.DeactivateTeamMembersReq
		setupMocks    func(*mockUserRepositoryForDeactivate, *mockPrReviewersRepositoryForDeactivate, *mockTeamRepositoryForDeactivate)
		expectedRes   *domain.DeactivateTeamMembersRes
		expectedError error
	}{
		{
			name: "successful deactivation without PRs",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "backend",
				UserIDs:  []string{"user1", "user2"},
			},
			setupMocks: func(userRepo *mockUserRepositoryForDeactivate, prReviewersRepo *mockPrReviewersRepositoryForDeactivate, teamRepo *mockTeamRepositoryForDeactivate) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "user1", IsActive: true},
							{UserID: "user2", IsActive: true},
							{UserID: "user3", IsActive: true},
						},
					}, nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
				teamRepo.deactivateTeamMembersFunc = func(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
					return []string{"user1", "user2"}, nil
				}
			},
			expectedRes: &domain.DeactivateTeamMembersRes{
				DeactivatedUserIDs: []string{"user1", "user2"},
			},
			expectedError: nil,
		},
		{
			name: "deactivation with PR reassignment",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "backend",
				UserIDs:  []string{"reviewer1"},
			},
			setupMocks: func(userRepo *mockUserRepositoryForDeactivate, prReviewersRepo *mockPrReviewersRepositoryForDeactivate, teamRepo *mockTeamRepositoryForDeactivate) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "reviewer1", IsActive: true},
							{UserID: "reviewer2", IsActive: true},
							{UserID: "author1", IsActive: true},
						},
					}, nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					if userID == "reviewer1" {
						return []domain.PullRequestShort{
							{PullRequestID: "pr1", Status: domain.PRStatusOpen},
						}, nil
					}
					return []domain.PullRequestShort{}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer1"}, nil
				}
				teamRepo.deactivateTeamMembersFunc = func(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
					return []string{"reviewer1"}, nil
				}
			},
			expectedRes: &domain.DeactivateTeamMembersRes{
				DeactivatedUserIDs: []string{"reviewer1"},
				Reassignments: []domain.ReviewerReassignment{
					{PrID: "pr1", OldReviewerID: "reviewer1", NewReviewerID: "reviewer2"},
				},
			},
			expectedError: nil,
		},
		{
			name: "team not found",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "nonexistent",
				UserIDs:  []string{"user1"},
			},
			setupMocks: func(userRepo *mockUserRepositoryForDeactivate, prReviewersRepo *mockPrReviewersRepositoryForDeactivate, teamRepo *mockTeamRepositoryForDeactivate) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedRes:   nil,
			expectedError: domain.ErrNotFound,
		},
		{
			name: "user not member of team",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "backend",
				UserIDs:  []string{"user1", "user2"},
			},
			setupMocks: func(userRepo *mockUserRepositoryForDeactivate, prReviewersRepo *mockPrReviewersRepositoryForDeactivate, teamRepo *mockTeamRepositoryForDeactivate) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "user3", IsActive: true},
						},
					}, nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{UserID: userID, TeamName: "other-team"}, nil
				}
			},
			expectedRes:   nil,
			expectedError: fmt.Errorf("%w: user user1 is not a member of team backend", domain.ErrInvalidRequest),
		},
		{
			name: "deactivate all team members successfully",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "backend",
				UserIDs:  []string{"user1", "user2"},
			},
			setupMocks: func(userRepo *mockUserRepositoryForDeactivate, prReviewersRepo *mockPrReviewersRepositoryForDeactivate, teamRepo *mockTeamRepositoryForDeactivate) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "user1", IsActive: true},
							{UserID: "user2", IsActive: true},
						},
					}, nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
				teamRepo.deactivateTeamMembersFunc = func(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
					return []string{"user1", "user2"}, nil
				}
			},
			expectedRes: &domain.DeactivateTeamMembersRes{
				DeactivatedUserIDs: []string{"user1", "user2"},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepositoryForDeactivate{}
			prReviewersRepo := &mockPrReviewersRepositoryForDeactivate{}
			teamRepo := &mockTeamRepositoryForDeactivate{}
			tt.setupMocks(userRepo, prReviewersRepo, teamRepo)

			service := NewUserService(userRepo, prReviewersRepo, teamRepo)
			res, err := service.DeactivateTeamMembers(context.Background(), tt.req)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if res == nil {
					t.Error("expected result, got nil")
					return
				}
				if len(res.DeactivatedUserIDs) != len(tt.expectedRes.DeactivatedUserIDs) {
					t.Errorf("expected %d deactivated users, got %d", len(tt.expectedRes.DeactivatedUserIDs), len(res.DeactivatedUserIDs))
				}
			}
		})
	}
}
