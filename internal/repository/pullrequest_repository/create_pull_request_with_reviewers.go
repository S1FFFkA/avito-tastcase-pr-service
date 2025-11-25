package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"

	"github.com/lib/pq"
)

func (s *PullRequestStorage) CreatePullRequestWithReviewers(
	ctx context.Context,
	pr *domain.PullRequest,
	reviewerIDs []string,
	needMoreReviewers bool,
) error {
	operation := "CreatePullRequestWithReviewers"

	logger.LogTransactionStart(operation)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.LogTransactionRollback(operation, err)
		return err
	}
	defer func() {
		if err != nil {
			logger.LogTransactionRollback(operation, err)
			_ = tx.Rollback()
		}
	}()

	query := `INSERT INTO pull_requests (id, pull_requests_name, author_id, status, need_more_reviewers) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, string(pr.Status), needMoreReviewers)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrPRExists
		}
		logger.LogQueryError(query, err)
		return err
	}

	if len(reviewerIDs) > 0 {
		reviewerQuery := `INSERT INTO reviewers (pull_request_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())`
		for _, reviewerID := range reviewerIDs {
			_, err = tx.ExecContext(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
			if err != nil {
				logger.LogQueryError(reviewerQuery, err)
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logger.LogTransactionRollback(operation, err)
		return err
	}

	return nil
}
