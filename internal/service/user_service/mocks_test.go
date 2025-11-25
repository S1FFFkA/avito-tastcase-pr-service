package service

import (
	"context"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

// Общие моки для тестов user_service

type mockUserRepositoryForService struct {
	repository.UserRepositoryInterface
	getUserByIDFunc     func(ctx context.Context, userID string) (*domain.User, error)
	setUserIsActiveFunc func(ctx context.Context, userID string, isActive bool) error
}

func (m *mockUserRepositoryForService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserRepositoryForService) SetUserIsActive(ctx context.Context, userID string, isActive bool) error {
	if m.setUserIsActiveFunc != nil {
		return m.setUserIsActiveFunc(ctx, userID, isActive)
	}
	return nil
}

type mockPrReviewersRepositoryForService struct {
	repository.PrReviewersRepositoryInterface
	getPRsByReviewerFunc func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *mockPrReviewersRepositoryForService) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return []domain.PullRequestShort{}, nil
}

type mockTeamRepositoryForUserService struct {
	repository.TeamRepositoryInterface
}
