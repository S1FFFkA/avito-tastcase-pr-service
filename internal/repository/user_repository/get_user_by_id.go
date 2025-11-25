package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"errors"
)

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var username string
	var teamName string
	var isActive bool

	query := `
		SELECT u.username, t.team_name, u.is_active
		FROM users u
		LEFT JOIN teams t ON u.team_id = t.id
		WHERE u.id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&username, &teamName, &isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logger.LogQueryError(query, err)
		return nil, err
	}

	user := &domain.User{
		UserID:   userID,
		Username: username,
		TeamName: teamName,
		IsActive: isActive,
	}

	return user, nil
}
