package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *PullRequestStorage) MergePullRequest(ctx context.Context, prID string) error {
	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = COALESCE(merged_at, $2)
		WHERE id = $3 AND status != $1`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query, string(domain.PRStatusMerged), now, prID)
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
		statusQuery := `SELECT status FROM pull_requests WHERE id = $1`
		var currentStatus string
		err = s.db.QueryRowContext(ctx, statusQuery, prID).Scan(&currentStatus)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrNotFound
			}
			logger.LogQueryError(statusQuery, err)
			return err
		}

		if currentStatus == string(domain.PRStatusMerged) {
			return nil
		}

		return domain.ErrNotFound
	}

	return nil
}
