package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type mockPullRequestRepository struct {
	repository.PullRequestRepositoryInterface
	getPullRequestByIDFunc             func(ctx context.Context, prID string) (*domain.PullRequest, error)
	createPullRequestWithReviewersFunc func(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error
}

func (m *mockPullRequestRepository) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	if m.getPullRequestByIDFunc != nil {
		return m.getPullRequestByIDFunc(ctx, prID)
	}
	return nil, nil
}

func (m *mockPullRequestRepository) CreatePullRequestWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
	if m.createPullRequestWithReviewersFunc != nil {
		return m.createPullRequestWithReviewersFunc(ctx, pr, reviewerIDs, needMoreReviewers)
	}
	return nil
}

type mockPrReviewersRepository struct {
	repository.PrReviewersRepositoryInterface
	getPRsByReviewerFunc func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *mockPrReviewersRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return []domain.PullRequestShort{}, nil
}

type mockUserRepositoryForPR struct {
	repository.UserRepositoryInterface
	getUserByIDFunc func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockUserRepositoryForPR) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

type mockTeamRepositoryForPR struct {
	repository.TeamRepositoryInterface
	getTeamByNameFunc func(ctx context.Context, teamName string) (*domain.Team, error)
}

func (m *mockTeamRepositoryForPR) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.getTeamByNameFunc != nil {
		return m.getTeamByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func TestPullRequestService_CreatePullRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.CreatePullRequestReq
		setupMocks    func(*mockPullRequestRepository, *mockPrReviewersRepository, *mockUserRepositoryForPR, *mockTeamRepositoryForPR)
		expectedError error
	}{
		{
			name: "successful creation",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-1",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "author-1",
						Username: "author1",
						TeamName: "backend",
						IsActive: true,
					}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "author-1", Username: "author1", IsActive: true},
							{UserID: "reviewer-1", Username: "reviewer1", IsActive: true},
							{UserID: "reviewer-2", Username: "reviewer2", IsActive: true},
						},
					}, nil
				}
				prRepo.createPullRequestWithReviewersFunc = func(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
					return nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "PR already exists",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-1",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{PullRequestID: "pr-1"}, nil
				}
				// Моки для userRepo и teamRepo не нужны, так как код должен вернуть ошибку до их вызова
			},
			expectedError: domain.ErrPRExists,
		},
		{
			name: "author not found",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-1",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedError: domain.ErrNotFound,
		},
		{
			name: "team not found",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-1",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "author-1",
						TeamName: "backend",
					}, nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedError: domain.ErrNotFound,
		},
		{
			name: "create PR with insufficient reviewers",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-2",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "author-1",
						TeamName: "backend",
					}, nil
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
				prRepo.createPullRequestWithReviewersFunc = func(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
					return nil
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error creating PR in database",
			req: &domain.CreatePullRequestReq{
				PullRequestID:   "pr-3",
				PullRequestName: "test-pr",
				AuthorID:        "author-1",
			},
			setupMocks: func(prRepo *mockPullRequestRepository, prReviewersRepo *mockPrReviewersRepository, userRepo *mockUserRepositoryForPR, teamRepo *mockTeamRepositoryForPR) {
				prRepo.getPullRequestByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, domain.ErrNotFound
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "author-1",
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
				prRepo.createPullRequestWithReviewersFunc = func(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
					return errors.New("database error")
				}
				prReviewersRepo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &mockPullRequestRepository{}
			prReviewersRepo := &mockPrReviewersRepository{}
			userRepo := &mockUserRepositoryForPR{}
			teamRepo := &mockTeamRepositoryForPR{}
			tt.setupMocks(prRepo, prReviewersRepo, userRepo, teamRepo)

			service := NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)
			_, err := service.CreatePullRequest(context.Background(), tt.req)

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
				}
			}
		})
	}
}
