package mocks

import (
	"context"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type MockPullRequestRepository struct {
	repository.PullRequestRepositoryInterface
	GetPullRequestByIDFunc             func(ctx context.Context, prID string) (*domain.PullRequest, error)
	CreatePullRequestWithReviewersFunc func(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error
	MergePullRequestFunc               func(ctx context.Context, prID string) error
	SetNeedMoreReviewersFunc           func(ctx context.Context, prID string, needMore bool) error
}

func (m *MockPullRequestRepository) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	if m.GetPullRequestByIDFunc != nil {
		return m.GetPullRequestByIDFunc(ctx, prID)
	}
	return nil, nil
}

func (m *MockPullRequestRepository) CreatePullRequestWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
	if m.CreatePullRequestWithReviewersFunc != nil {
		return m.CreatePullRequestWithReviewersFunc(ctx, pr, reviewerIDs, needMoreReviewers)
	}
	return nil
}

func (m *MockPullRequestRepository) MergePullRequest(ctx context.Context, prID string) error {
	if m.MergePullRequestFunc != nil {
		return m.MergePullRequestFunc(ctx, prID)
	}
	return nil
}

func (m *MockPullRequestRepository) SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error {
	if m.SetNeedMoreReviewersFunc != nil {
		return m.SetNeedMoreReviewersFunc(ctx, prID, needMore)
	}
	return nil
}
