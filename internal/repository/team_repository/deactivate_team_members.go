package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"

	"github.com/lib/pq"
)

func (s *TeamStorage) DeactivateTeamMembers(
	ctx context.Context,
	teamName string,
	userIDs []string,
	reassignments []domain.ReviewerReassignment,
) ([]string, error) {
	operation := "DeactivateTeamMembers"

	logger.LogTransactionStart(operation)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.LogTransactionRollback(operation, err)
		return nil, err
	}
	defer func() {
		if err != nil {
			logger.LogTransactionRollback(operation, err)
			_ = tx.Rollback()
		}
	}()

	query := `
		UPDATE users u
		SET is_active = false
		FROM teams t
		WHERE u.team_id = t.id AND t.team_name = $1`

	args := make([]interface{}, 0, 2)
	args = append(args, teamName)

	if len(userIDs) > 0 {
		query += ` AND u.id = ANY($2)`
		args = append(args, pq.Array(userIDs))
	}

	query += ` RETURNING u.id`

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}
	defer rows.Close()

	capacity := 10
	if len(userIDs) > 0 {
		capacity = len(userIDs)
	}
	deactivatedIDs := make([]string, 0, capacity)
	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			logger.LogQueryError(query, err)
			return nil, err
		}
		deactivatedIDs = append(deactivatedIDs, userID)
	}

	if err = rows.Err(); err != nil {
		logger.LogQueryError(query, err)
		return nil, err
	}

	for _, reassignment := range reassignments {
		deleteQuery := `DELETE FROM reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
		_, err = tx.ExecContext(ctx, deleteQuery, reassignment.PrID, reassignment.OldReviewerID)
		if err != nil {
			logger.LogQueryError(deleteQuery, err)
			return nil, err
		}

		if reassignment.NewReviewerID != "" {
			insertQuery := `INSERT INTO reviewers (pull_request_id, reviewer_id, assigned_at) VALUES ($1, $2, NOW())`
			_, err = tx.ExecContext(ctx, insertQuery, reassignment.PrID, reassignment.NewReviewerID)
			if err != nil {
				logger.LogQueryError(insertQuery, err)
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logger.LogTransactionRollback(operation, err)
		return nil, err
	}

	return deactivatedIDs, nil
}
