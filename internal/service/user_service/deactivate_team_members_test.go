package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	logger.InitLogger()
}

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) SetUserIsActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

type MockPrReviewersRepository struct {
	mock.Mock
}

func (m *MockPrReviewersRepository) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPrReviewersRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PullRequestShort), args.Error(1)
}

func (m *MockPrReviewersRepository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	args := m.Called(ctx, prID, oldReviewerID, newReviewerID)
	return args.Error(0)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
	args := m.Called(ctx, teamName, members)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockTeamRepository) DeactivateTeamMembers(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
	args := m.Called(ctx, teamName, userIDs, reassignments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestUserServiceImpl_DeactivateTeamMembers(t *testing.T) {
	tests := []struct {
		name                 string
		req                  *domain.DeactivateTeamMembersReq
		setupMocks           func(*MockTeamRepository, *MockPrReviewersRepository, *MockUserRepository)
		wantErr              error
		wantDeactivatedCount int
	}{
		{
			name: "successful deactivation",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{"user1"},
			},
			setupMocks: func(teamRepo *MockTeamRepository, prRepo *MockPrReviewersRepository, userRepo *MockUserRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{UserID: "user1", IsActive: true},
						{UserID: "user2", IsActive: true},
					},
				}, nil)
				prRepo.On("GetPRsByReviewer", mock.Anything, "user1").Return([]domain.PullRequestShort{}, nil)
				teamRepo.On("DeactivateTeamMembers", mock.Anything, "team1", []string{"user1"}, mock.Anything).Return([]string{"user1"}, nil)
			},
			wantErr:              nil,
			wantDeactivatedCount: 1,
		},
		{
			name: "team not found",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "nonexistent",
				UserIDs:  []string{"user1"},
			},
			setupMocks: func(teamRepo *MockTeamRepository, prRepo *MockPrReviewersRepository, userRepo *MockUserRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "nonexistent").Return(nil, domain.ErrNotFound)
			},
			wantErr:              domain.ErrNotFound,
			wantDeactivatedCount: 0,
		},
		{
			name: "user not a member of team",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{"user999"},
			},
			setupMocks: func(teamRepo *MockTeamRepository, prRepo *MockPrReviewersRepository, userRepo *MockUserRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{UserID: "user1", IsActive: true},
					},
				}, nil)
			},
			wantErr:              domain.ErrInvalidRequest,
			wantDeactivatedCount: 0,
		},
		{
			name: "empty user_ids list",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{},
			},
			setupMocks: func(teamRepo *MockTeamRepository, prRepo *MockPrReviewersRepository, userRepo *MockUserRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{UserID: "user1", IsActive: true},
					},
				}, nil)
			},
			wantErr:              domain.ErrInvalidRequest,
			wantDeactivatedCount: 0,
		},
		{
			name: "all team members in user_ids",
			req: &domain.DeactivateTeamMembersReq{
				TeamName: "team1",
				UserIDs:  []string{"user1", "user2"},
			},
			setupMocks: func(teamRepo *MockTeamRepository, prRepo *MockPrReviewersRepository, userRepo *MockUserRepository) {
				teamRepo.On("GetTeamByName", mock.Anything, "team1").Return(&domain.Team{
					TeamName: "team1",
					Members: []domain.TeamMember{
						{UserID: "user1", IsActive: true},
						{UserID: "user2", IsActive: true},
					},
				}, nil)
			},
			wantErr:              domain.ErrInvalidRequest,
			wantDeactivatedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamRepo := new(MockTeamRepository)
			prRepo := new(MockPrReviewersRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(teamRepo, prRepo, userRepo)

			service := &UserServiceImpl{
				teamRepo:        teamRepo,
				prReviewersRepo: prRepo,
				userRepo:        userRepo,
			}

			result, err := service.DeactivateTeamMembers(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.DeactivatedUserIDs, tt.wantDeactivatedCount)
			}

			teamRepo.AssertExpectations(t)
			prRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}
