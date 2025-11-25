package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type mockPullRequestRepositoryForReassign struct {
	repository.PullRequestRepositoryInterface
	getPullRequestByIDFunc   func(ctx context.Context, prID string) (*domain.PullRequest, error)
	setNeedMoreReviewersFunc func(ctx context.Context, prID string, needMore bool) error
}

func (m *mockPullRequestRepositoryForReassign) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	if m.getPullRequestByIDFunc != nil {
		return m.getPullRequestByIDFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPullRequestRepositoryForReassign) SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error {
	if m.setNeedMoreReviewersFunc != nil {
		return m.setNeedMoreReviewersFunc(ctx, prID, needMore)
	}
	return nil
}

type mockPrReviewersRepositoryForReassign struct {
	repository.PrReviewersRepositoryInterface
	getAssignedReviewersFunc func(ctx context.Context, prID string) ([]string, error)
	reassignReviewerFunc     func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	getPRsByReviewerFunc     func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *mockPrReviewersRepositoryForReassign) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	if m.getAssignedReviewersFunc != nil {
		return m.getAssignedReviewersFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPrReviewersRepositoryForReassign) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	if m.reassignReviewerFunc != nil {
		return m.reassignReviewerFunc(ctx, prID, oldReviewerID, newReviewerID)
	}
	return nil
}

func (m *mockPrReviewersRepositoryForReassign) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return []domain.PullRequestShort{}, nil
}

func TestPullRequestService_ReassignReviewer(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.ReassignReviewerReq
		setupMocks    func(*mockPullRequestRepositoryForReassign, *mockPrReviewersRepositoryForReassign, *mockUserRepositoryForPR, *mockTeamRepositoryForPR)
		expectedPR    *domain.PullRequest
		expectedNewID string
		expectedError error
	}{
		{
			name: "successful reassignment",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						AuthorID:      "author-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					if prID == "pr-1" {
						return []string{"reviewer-1", "reviewer-2"}, nil
					}
					return []string{"reviewer-2", "reviewer-3"}, nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "reviewer-1",
						TeamName: "backend",
					}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "author-1", IsActive: true},
							{UserID: "reviewer-1", IsActive: true},
							{UserID: "reviewer-2", IsActive: true},
							{UserID: "reviewer-3", IsActive: true},
						},
					}, nil
				}
				prReviewersRepo.reassignReviewerFunc = func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
					return nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedPR: &domain.PullRequest{
				PullRequestID:     "pr-1",
				AssignedReviewers: []string{"reviewer-2", "reviewer-3"},
			},
			expectedNewID: "reviewer-3",
			expectedError: nil,
		},
		{
			name: "PR already merged",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						Status:        domain.PRStatusMerged,
					}, nil
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: domain.ErrPRMerged,
		},
		{
			name: "reviewer not assigned",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-3",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1", "reviewer-2"}, nil
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: domain.ErrNotAssigned,
		},
		{
			name: "no candidate available",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						AuthorID:      "author-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1", "reviewer-2"}, nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "reviewer-1",
						TeamName: "backend",
					}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "author-1", IsActive: true},
							{UserID: "reviewer-1", IsActive: true},
							{UserID: "reviewer-2", IsActive: true},
						},
					}, nil
				}
				prRepo.setNeedMoreReviewersFunc = func(ctx context.Context, prID string, needMore bool) error {
					return nil
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: domain.ErrNoCandidate,
		},
		{
			name: "reassign when only author available",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						AuthorID:      "author-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1"}, nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{UserID: userID, TeamName: "backend"}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "author-1", IsActive: true},
							{UserID: "reviewer-1", IsActive: true},
						},
					}, nil
				}
				prRepo.setNeedMoreReviewersFunc = func(ctx context.Context, prID string, needMore bool) error {
					return nil
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: domain.ErrNoCandidate,
		},
		{
			name: "error getting PR status",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, errors.New("database error")
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: errors.New("database error"),
		},
		{
			name: "error during reassignment",
			req: &domain.ReassignReviewerReq{
				PullRequestID: "pr-1",
				OldUserID:     "reviewer-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepositoryForReassign, prReviewersRepo *mockPrReviewersRepositoryForReassign, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						PullRequestID: "pr-1",
						AuthorID:      "author-1",
						Status:        domain.PRStatusOpen,
					}, nil
				}
				prReviewersRepo.getAssignedReviewersFunc = func(ctx context.Context, prID string) ([]string, error) {
					return []string{"reviewer-1"}, nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{UserID: userID, TeamName: "backend"}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "author-1", IsActive: true},
							{UserID: "reviewer-1", IsActive: true},
							{UserID: "reviewer-2", IsActive: true},
						},
					}, nil
				}
				prReviewersRepo.reassignReviewerFunc = func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
					return errors.New("reassignment failed")
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedPR:    nil,
			expectedNewID: "",
			expectedError: errors.New("reassignment failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &mockPullRequestRepositoryForReassign{}
			prReviewersRepo := &mockPrReviewersRepositoryForReassign{}
			userRepo := &mockUserRepositoryForPR{}
			teamRepo := &mockTeamRepositoryForPR{}
			tt.setupMocks(prRepo, prReviewersRepo, userRepo, teamRepo)

			service := NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)
			pr, newReviewerID, err := service.ReassignReviewer(context.Background(), tt.req)

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
				if newReviewerID != tt.expectedNewID {
					t.Errorf("expected new reviewer ID %s, got %s", tt.expectedNewID, newReviewerID)
				}
			}
		})
	}
}
