package service

import (
	"context"
	"errors"
	"testing"

	"AVITOSAMPISHU/internal/domain"
)

func TestTeamService_GetTeam(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		setupMocks    func(*mockTeamRepository, *mockUserRepository)
		expectedTeam  *domain.Team
		expectedError error
	}{
		{
			name:     "successful get",
			teamName: "backend",
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "user-1", Username: "user1", IsActive: true},
							{UserID: "user-2", Username: "user2", IsActive: true},
						},
					}, nil
				}
			},
			expectedTeam: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
					{UserID: "user-2", Username: "user2", IsActive: true},
				},
			},
			expectedError: nil,
		},
		{
			name:     "team not found",
			teamName: "nonexistent",
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, domain.ErrNotFound
				}
			},
			expectedTeam:  nil,
			expectedError: domain.ErrNotFound,
		},
		{
			name:     "database error",
			teamName: "backend",
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, errors.New("database error")
				}
			},
			expectedTeam:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name:     "get team with empty members",
			teamName: "empty-team",
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "empty-team",
						Members:  []domain.TeamMember{},
					}, nil
				}
			},
			expectedTeam: &domain.Team{
				TeamName: "empty-team",
				Members:  []domain.TeamMember{},
			},
			expectedError: nil,
		},
		{
			name:     "get team with inactive members",
			teamName: "backend",
			setupMocks: func(teamRepo *mockTeamRepository, userRepo *mockUserRepository) {
				teamRepo.getTeamByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						TeamName: "backend",
						Members: []domain.TeamMember{
							{UserID: "user-1", Username: "user1", IsActive: true},
							{UserID: "user-2", Username: "user2", IsActive: false},
						},
					}, nil
				}
			},
			expectedTeam: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserID: "user-1", Username: "user1", IsActive: true},
					{UserID: "user-2", Username: "user2", IsActive: false},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamRepo := &mockTeamRepository{}
			userRepo := &mockUserRepository{}
			tt.setupMocks(teamRepo, userRepo)

			service := NewTeamService(teamRepo, userRepo)
			team, err := service.GetTeam(context.Background(), tt.teamName)

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
				if team.TeamName != tt.expectedTeam.TeamName {
					t.Errorf("expected team name %s, got %s", tt.expectedTeam.TeamName, team.TeamName)
				}
				if len(team.Members) != len(tt.expectedTeam.Members) {
					t.Errorf("expected %d members, got %d", len(tt.expectedTeam.Members), len(team.Members))
				}
			}
		})
	}
}
