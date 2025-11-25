package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"time"
)

func (s *PullRequestServiceImpl) MergePullRequest(
	ctx context.Context,
	req *domain.MergePullRequestReq,
) (*domain.PullRequest, error) {
	start := time.Now()
	operation := "MergePullRequest"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"pr_id": req.PullRequestID,
	})

	pr, err := s.prRepo.GetPullRequestByID(ctx, req.PullRequestID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, err
	}

	if pr.Status == domain.PRStatusOpen {
		if err := s.prRepo.MergePullRequest(ctx, req.PullRequestID); err != nil {
			logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
				"pr_id": req.PullRequestID,
				"error": err.Error(),
			})
			return nil, err
		}

		now := time.Now()
		pr.Status = domain.PRStatusMerged
		pr.MergedAt = &now
	}

	reviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, err
	}

	pr.AssignedReviewers = reviewers

	if pr.Status == domain.PRStatusMerged {
		helpers.UpdateReviewerLoadMetrics(ctx, s.prReviewersRepo, reviewers)
	}

	logger.LogBusinessTransactionEnd(operation, time.Since(start), true, map[string]interface{}{
		"pr_id": req.PullRequestID,
	})
	logger.LogCriticalEvent("pull_request_merged", map[string]interface{}{
		"pr_id": req.PullRequestID,
	})

	return pr, nil
}
