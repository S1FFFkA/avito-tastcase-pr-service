package mocks

import (
	"context"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type MockPrReviewersRepository struct {
	repository.PrReviewersRepositoryInterface
	GetAssignedReviewersFunc func(ctx context.Context, prID string) ([]string, error)
	GetPRsByReviewerFunc     func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	ReassignReviewerFunc     func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}

func (m *MockPrReviewersRepository) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	if m.GetAssignedReviewersFunc != nil {
		return m.GetAssignedReviewersFunc(ctx, prID)
	}
	return nil, nil
}

func (m *MockPrReviewersRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.GetPRsByReviewerFunc != nil {
		return m.GetPRsByReviewerFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPrReviewersRepository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	if m.ReassignReviewerFunc != nil {
		return m.ReassignReviewerFunc(ctx, prID, oldReviewerID, newReviewerID)
	}
	return nil
}
