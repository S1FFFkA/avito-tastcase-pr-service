package mocks

import (
	"context"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
)

type MockUserRepository struct {
	repository.UserRepositoryInterface
	GetUserByIDFunc     func(ctx context.Context, userID string) (*domain.User, error)
	SetUserIsActiveFunc func(ctx context.Context, userID string, isActive bool) error
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepository) SetUserIsActive(ctx context.Context, userID string, isActive bool) error {
	if m.SetUserIsActiveFunc != nil {
		return m.SetUserIsActiveFunc(ctx, userID, isActive)
	}
	return nil
}
