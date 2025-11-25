package mocks

import (
	"context"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"

	"github.com/google/uuid"
)

type MockTeamRepository struct {
	repository.TeamRepositoryInterface
	GetTeamByNameFunc         func(ctx context.Context, teamName string) (*domain.Team, error)
	CreateTeamWithMembersFunc func(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error)
	DeactivateTeamMembersFunc func(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error)
}

func (m *MockTeamRepository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.GetTeamByNameFunc != nil {
		return m.GetTeamByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func (m *MockTeamRepository) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
	if m.CreateTeamWithMembersFunc != nil {
		return m.CreateTeamWithMembersFunc(ctx, teamName, members)
	}
	return uuid.Nil, nil
}

func (m *MockTeamRepository) DeactivateTeamMembers(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error) {
	if m.DeactivateTeamMembersFunc != nil {
		return m.DeactivateTeamMembersFunc(ctx, teamName, userIDs, reassignments)
	}
	return nil, nil
}
