package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type mockPullRequestRepositoryForMerge struct {
	repository.PullRequestRepositoryInterface
	getPullRequestByIDFunc func(ctx context.Context, prID string) (*domain.PullRequest, error)
	mergePullRequestFunc   func(ctx context.Context, prID string) error
}

func (m *mockPullRequestRepositoryForMerge) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	if m.getPullRequestByIDFunc != nil {
		return m.getPullRequestByIDFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPullRequestRepositoryForMerge) MergePullRequest(ctx context.Context, prID string) error {
	if m.mergePullRequestFunc != nil {
		return m.mergePullRequestFunc(ctx, prID)
	}
	return nil
}

type mockPrReviewersRepositoryForMerge struct {
	repository.PrReviewersRepositoryInterface
	getAssignedReviewersFunc func(ctx context.Context, prID string) ([]string, error)
	getPRsByReviewerFunc     func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *mockPrReviewersRepositoryForMerge) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	if m.getAssignedReviewersFunc != nil {
		return m.getAssignedReviewersFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPrReviewersRepositoryForMerge) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return []domain.PullRequestShort{}, nil
}

func TestPullRequestService_MergePullRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.MergePullRequestReq
		setupMocks    func(*mockPullRequestRepositoryForMerge, *mockPrReviewersRepositoryForMerge, *mockUserRepositoryForPR, *mockTeamRepositoryForPR)
		expectedPR    *domain.PullRequest
		expectedError error
	}{
		{
			name: "successful merge",
			req: &domain.MergePullRequestReq{
				PullRequestID: "pr-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForMerge, prReviewersRepo *mockPrReviewersRepositoryForMerge, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prRepo.mergePullRequestFunc = func(ctx context.Context, prID string) error {
					return nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1", "reviewer-2"}, nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-1",
				Status:            domain.PRStatusMerged,
				AssignedReviewers: []string{"reviewer-1", "reviewer-2"},
			},
			expectedError: nil,
		},
		{
			name: "already merged (idempotent)",
			req: &domain.MergePullRequestReq{
				PullRequestID: "pr-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForMerge, prReviewersRepo *mockPrReviewersRepositoryForMerge, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						Status:        domain.PRStatusMerged,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1"}, nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-1",
				Status:            domain.PRStatusMerged,
				AssignedReviewers: []string{"reviewer-1"},
			},
			expectedError: nil,
		},
		{
			name: "PR not found",
			req: &domain.MergePullRequestReq{
				PullRequestID: "nonexistent",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForMerge, prReviewersRepo *mockPrReviewersRepositoryForMerge, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedPR:    nil,
			expectedError: domain.ErrNotFound,
		},
		{
			name: "database error on merge",
			req: &domain.MergePullRequestReq{
				PullRequestID: "pr-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForMerge, prReviewersRepo *mockPrReviewersRepositoryForMerge, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prRepo.mergePullRequestFunc = func(ctx context.Context, prID string) error {
					return errors.New("database error")
				}
			},
			expectedPR:    nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &mockPullRequestRepositoryForMerge{}
			prReviewersRepo := &mockPrReviewersRepositoryForMerge{}
			userRepo := &mockUserRepositoryForPR{}
			teamRepo := &mockTeamRepositoryForPR{}
			tt.setupMocks(prRepo, prReviewersRepo, userRepo, teamRepo)

			service := NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)
			pr, err := service.MergePullRequest(context.Background(), tt.req)

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
				if pr.PullRequestID != tt.expectedPR.PullRequestID {
					t.Errorf("expected PR ID %s, got %s", tt.expectedPR.PullRequestID, pr.PullRequestID)
				}
				if pr.Status != tt.expectedPR.Status {
					t.Errorf("expected status %s, got %s", tt.expectedPR.Status, pr.Status)
				}
			}
		})
	}
}
