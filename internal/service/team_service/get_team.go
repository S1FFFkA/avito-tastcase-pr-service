package service

import (
	"AVITOSAMPISHU/internal/domain"
	"context"
)

func (s *TeamServiceImpl) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return team, nil
}
