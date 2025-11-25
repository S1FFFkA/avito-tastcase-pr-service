package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"time"
)

func (s *TeamServiceImpl) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	start := time.Now()
	operation := "CreateTeam"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"team_name":     team.TeamName,
		"members_count": len(team.Members),
	})

	if _, err := s.teamRepo.CreateTeamWithMembers(ctx, team.TeamName, team.Members); err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": team.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	createdTeam, err := s.teamRepo.GetTeamByName(ctx, team.TeamName)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": team.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	logger.LogBusinessTransactionEnd(operation, time.Since(start), true, map[string]interface{}{
		"team_name": createdTeam.TeamName,
	})
	logger.LogCriticalEvent("team_created", map[string]interface{}{
		"team_name": createdTeam.TeamName,
	})

	return createdTeam, nil
}
