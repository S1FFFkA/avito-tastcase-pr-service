package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
)

func (s *TeamStorage) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	query := `
		SELECT u.id, u.username, u.is_active
		FROM teams t
		LEFT JOIN users u ON t.id = u.team_id
		WHERE t.team_name = $1
		ORDER BY u.username`

	rows, err := s.db.QueryContext(ctx, query, teamName)
	if err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0, 10)
	var teamExists bool

	for rows.Next() {
		var userID sql.NullString
		var username sql.NullString
		var isActive sql.NullBool

		if err = rows.Scan(&userID, &username, &isActive); err != nil {
			logger.LogQueryError(query, err)
			return nil, err
		}

		teamExists = true

		if userID.Valid {
			members = append(members, domain.TeamMember{
				UserID:   userID.String,
				Username: username.String,
				IsActive: isActive.Bool,
			})
		}
	}

	if err = rows.Err(); err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}

	if !teamExists || len(members) == 0 {
		return nil, domain.ErrNotFound
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
