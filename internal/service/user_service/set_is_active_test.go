package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
)

func TestUserService_SetIsActive(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.SetIsActiveRequest
		setupMocks    func(*mockUserRepositoryForService, *mockPrReviewersRepositoryForService, *mockTeamRepositoryForUserService)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name: "successful activation",
			req: &domain.SetIsActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "user-1",
						Username: "user1",
						IsActive: false,
					}, nil
				}
				userRepo.setUserIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return nil
				}
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						UserID:   "user-1",
						Username: "user1",
						IsActive: true,
					}, nil
				}
			},
			expectedUser: &domain.User{
				UserID:   "user-1",
				Username: "user1",
				IsActive: true,
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			req: &domain.SetIsActiveRequest{
				UserID:   "nonexistent",
				IsActive: true,
			},
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedUser:  nil,
			expectedError: domain.ErrNotFound,
		},
		{
			name: "database error on update",
			req: &domain.SetIsActiveRequest{
				UserID:   "user-1",
				IsActive: true,
			},
			setupMocks: func(userRepo *mockUserRepositoryForService, prReviewersRepo *mockPrReviewersRepositoryForService, teamRepo *mockTeamRepositoryForUserService) {
				userRepo.getUserByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{UserID: "user-1"}, nil
				}
				userRepo.setUserIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return errors.New("database error")
				}
			},
			expectedUser:  nil,
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
			user, err := service.SetIsActive(context.Background(), tt.req)

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
				if user.UserID != tt.expectedUser.UserID {
					t.Errorf("expected user ID %s, got %s", tt.expectedUser.UserID, user.UserID)
				}
				if user.IsActive != tt.expectedUser.IsActive {
					t.Errorf("expected is_active %v, got %v", tt.expectedUser.IsActive, user.IsActive)
				}
			}
		})
	}
}
