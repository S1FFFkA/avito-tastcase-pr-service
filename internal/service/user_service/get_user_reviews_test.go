package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
)

func TestUserService_GetUserReviews(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*mockUserRepositoryForService, *mockPrReviewersRepositoryForService, *mockTeamRepositoryForUserService)
		expectedPRs   []domain.PullRequestShort
		expectedError error
	}{
		{
			name:   "successful get",
			userID: "user-1",
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{
						{PullRequestID: "pr-1", PullRequestName: "pr1", AuthorID: "author-1", Status: domain.PRStatusOpen},
						{PullRequestID: "pr-2", PullRequestName: "pr2", AuthorID: "author-2", Status: domain.PRStatusMerged},
					}, nil
				}
			},
			expectedPRs: []domain.PullRequestShort{
				{PullRequestID: "pr-1", PullRequestName: "pr1", AuthorID: "author-1", Status: domain.PRStatusOpen},
				{PullRequestID: "pr-2", PullRequestName: "pr2", AuthorID: "author-2", Status: domain.PRStatusMerged},
			},
			expectedError: nil,
		},
		{
			name:   "no PRs found",
			userID: "user-1",
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedPRs:   []domain.PullRequestShort{},
			expectedError: nil,
		},
		{
			name:   "database error",
			userID: "user-1",
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return nil, errors.New("database error")
				}
			},
			expectedPRs:   nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepositoryForService{}
			prReviewersRepo := &mockPrReviewersRepositoryForService{}
			teamRepo := &mockTeamRepositoryForUserService{}
			tt.setupMocks(userRepo, prReviewersRepo, teamRepo)

			service := NewUserService(userRepo, prReviewersRepo, teamRepo)
			prs, err := service.GetUserReviews(context.Background(), tt.userID)

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
				if len(prs) != len(tt.expectedPRs) {
					t.Errorf("expected %d PRs, got %d", len(tt.expectedPRs), len(prs))
				}
			}
		})
	}
}
