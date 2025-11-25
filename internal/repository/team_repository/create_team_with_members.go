package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (s *TeamStorage) CreateTeamWithMembers(
	ctx context.Context,
	teamName string,
	members []domain.TeamMember,
) (uuid.UUID, error) {
	operation := "CreateTeamWithMembers"

	logger.LogTransactionStart(operation)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.LogTransactionRollback(operation, err)
		return uuid.Nil, err
	}
	defer func() {
		if err != nil {
			logger.LogTransactionRollback(operation, err)
			_ = tx.Rollback()
		}
	}()

	var teamID uuid.UUID
	query := `INSERT INTO teams (team_name) VALUES ($1) RETURNING id`
	err = tx.QueryRowContext(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return uuid.Nil, domain.ErrTeamExists
		}
		logger.LogQueryError(query, err)
		return uuid.Nil, err
	}

	userQuery := `INSERT INTO users (id, username, team_id, is_active) VALUES ($1, $2, $3, $4)`
	for _, member := range members {
		_, err = tx.ExecContext(ctx, userQuery, member.UserID, member.Username, teamID, member.IsActive)
		if err != nil {
			logger.LogQueryError(userQuery, err)
			return uuid.Nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.LogTransactionRollback(operation, err)
		return uuid.Nil, err
	}

	return teamID, nil
}
