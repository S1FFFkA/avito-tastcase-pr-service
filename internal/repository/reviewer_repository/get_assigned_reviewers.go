package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
)

func (s *PrReviewersStorage) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1 ORDER BY assigned_at`

	rows, err := s.db.QueryContext(ctx, query, prID)
	if err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0, domain.MaxReviewersCount)
	for rows.Next() {
		var reviewerID string
		if err = rows.Scan(&reviewerID); err != nil {
			logger.LogQueryError(query, err)
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}

	return reviewers, nil
}
