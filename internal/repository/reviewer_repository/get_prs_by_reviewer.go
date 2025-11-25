package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
)

func (s *PrReviewersStorage) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT pr.id, pr.pull_requests_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN reviewers r ON pr.id = r.pull_request_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}
	defer rows.Close()

	prs := make([]domain.PullRequestShort, 0, 20)
	for rows.Next() {
		var prID string
		var name string
		var authorID string
		var status string

		if err = rows.Scan(&prID, &name, &authorID, &status); err != nil {
			logger.LogQueryError(query, err)
			return nil, err
		}

		prs = append(prs, domain.PullRequestShort{
			PullRequestID:   prID,
			PullRequestName: name,
			AuthorID:        authorID,
			Status:          domain.PRStatus(status),
		})
	}

	if err = rows.Err(); err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}

	return prs, nil
}
