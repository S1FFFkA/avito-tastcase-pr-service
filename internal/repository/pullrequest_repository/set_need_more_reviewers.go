package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
)

func (s *PullRequestStorage) SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error {
	query := `UPDATE pull_requests SET need_more_reviewers = $1 WHERE id = $2`

	result, err := s.db.ExecContext(ctx, query, needMore, prID)
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
