package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"

	"github.com/google/uuid"
)

type mockTeamRepository struct {
	repository.TeamRepositoryInterface
	getTeamByNameFunc         func(ctx context.Context, teamName string) (*domain.Team, error)
	createTeamWithMembersFunc func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error)
}

func (m *mockTeamRepository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.getTeamByNameFunc != nil {
		return m.getTeamByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func (m *mockTeamRepository) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
	if m.createTeamWithMembersFunc != nil {
		return m.createTeamWithMembersFunc(ctx, teamName, members)
	}
	return uuid.Nil, nil
}

type mockUserRepository struct {
	repository.UserRepositoryInterface
}

func TestTeamService_CreateTeam(t *testing.T) {
	tests := []struct {
		name          string
		team          *domain.Team
		setupMocks    func(*mockTeamRepository, *mockUserRepository)
		expectedError error
	}{
		{
			name: "successful creation",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{TeamName: teamName, Members: []domain.TeamMember{}}, nil
				}
				teamRepo.createTeamWithMembersFunc = func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
					return uuid.New(), nil
				}
			},
			expectedError: nil,
		},
		{
			name: "team already exists",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.createTeamWithMembersFunc = func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
					return uuid.Nil, domain.ErrTeamExists
				}
			},
			expectedError: domain.ErrTeamExists,
		},
		{
			name: "database error",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "create team with multiple members",
			team: &domain.Team{
				TeamName: "frontend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
					{UserID: "user-2", Username: "user2", IsActive: true},
					{UserID: "user-3", Username: "user3", IsActive: false},
				},
			},
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: teamName,
						Members: []domain.TeamMember{
							{UserID: "user-1", Username: "user1", IsActive: true},
							{UserID: "user-2", Username: "user2", IsActive: true},
							{UserID: "user-3", Username: "user3", IsActive: false},
						},
					}, nil
				}
				teamRepo.createTeamWithMembersFunc = func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
					return uuid.New(), nil
				}
			},
			expectedError: nil,
		},
		{
			name: "create team error on get after create",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
				},
			},
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.createTeamWithMembersFunc = func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
					return uuid.New(), nil
				}
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, errors.New("database error on get")
				}
			},
			expectedError: errors.New("database error on get"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamRepo := &mockTeamRepository{}
			userRepo := &mockUserRepository{}
			tt.setupMocks(teamRepo, userRepo)

			service := NewTeamService(teamRepo, userRepo)
			_, err := service.CreateTeam(context.Background(), tt.team)

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
