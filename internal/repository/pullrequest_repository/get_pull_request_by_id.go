package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *PullRequestStorage) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
		SELECT pull_requests_name, author_id, status, need_more_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE id = $1`

	var name string
	var authorID string
	var status string
	var needMoreReviewers bool
	var createdAt time.Time
	var mergedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, prID).Scan(&name, &authorID, &status, &needMoreReviewers, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logger.LogQueryError(query, err)
		return nil, err
	}

	reviewersQuery := `SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1`
	rows, err := s.db.QueryContext(ctx, reviewersQuery, prID)
	if err != nil {
		logger.LogQueryError(reviewersQuery, err)
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0, domain.MaxReviewersCount)
	for rows.Next() {
		var reviewerID string
		if err = rows.Scan(&reviewerID); err != nil {
			logger.LogQueryError(reviewersQuery, err)
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		logger.LogQueryError(reviewersQuery, err)
		return nil, err
	}

	var mergedAtPtr *time.Time
	if mergedAt.Valid {
		mergedAtPtr = &mergedAt.Time
	}

	return &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   name,
		AuthorID:          authorID,
		Status:            domain.PRStatus(status),
		AssignedReviewers: reviewers,
		NeedMoreReviewers: &needMoreReviewers,
		CreatedAt:         &createdAt,
		MergedAt:          mergedAtPtr,
	}, nil
}
