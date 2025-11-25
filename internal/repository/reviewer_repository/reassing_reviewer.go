package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"errors"
)

func (s *PrReviewersStorage) ReassignReviewer(
	ctx context.Context,
	prID,
	oldReviewerID,
	newReviewerID string,
) error {
	operation := "ReassignReviewer"

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

	statusQuery := `SELECT status FROM pull_requests WHERE id = $1`
	var prStatus string
	err = tx.QueryRowContext(ctx, statusQuery, prID).Scan(&prStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		logger.LogQueryError(statusQuery, err)
		return err
	}

	if prStatus == string(domain.PRStatusMerged) {
		return domain.ErrPRMerged
	}

	deleteQuery := `DELETE FROM reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	result, err := tx.ExecContext(ctx, deleteQuery, prID, oldReviewerID)
	if err != nil {
		logger.LogQueryError(deleteQuery, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.LogQueryError(deleteQuery, err)
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotAssigned
	}

	if newReviewerID != "" {
		insertQuery := `INSERT INTO reviewers (pull_request_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())`
		_, err = tx.ExecContext(ctx, insertQuery, prID, newReviewerID)
		if err != nil {
			logger.LogQueryError(insertQuery, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.LogTransactionRollback(operation, err)
		return err
	}

	return nil
}
