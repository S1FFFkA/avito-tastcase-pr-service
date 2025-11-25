package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
)

func (r *UserRepository) SetUserIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		logger.LogQueryError(query, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.LogQueryError(query, err)
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
