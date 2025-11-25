package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"time"
)

func (s *PullRequestServiceImpl) CreatePullRequest(
	ctx context.Context,
	req *domain.CreatePullRequestReq,
) (*domain.PullRequest, error) {
	start := time.Now()
	operation := "CreatePullRequest"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"pr_id":   req.PullRequestID,
		"pr_name": req.PullRequestName,
		"author":  req.AuthorID,
	})

	existingPR, err := s.prRepo.GetPullRequestByID(ctx, req.PullRequestID)
	if err == nil && existingPR != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id":  req.PullRequestID,
			"error":  domain.ErrPRExists.Error(),
			"reason": "pr_already_exists",
		})
		return nil, domain.ErrPRExists
	}
	if err != nil && err != domain.ErrNotFound {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id":  req.PullRequestID,
			"error":  err.Error(),
			"reason": "error_checking_pr",
		})
		return nil, err
	}

	author, err := s.userRepo.GetUserByID(ctx, req.AuthorID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id":  req.PullRequestID,
			"error":  err.Error(),
			"reason": "author_not_found",
		})
		return nil, err
	}

	team, err := s.teamRepo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id":  req.PullRequestID,
			"error":  err.Error(),
			"reason": "team_not_found",
		})
		return nil, err
	}

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestID:     req.PullRequestID,
		PullRequestName:   req.PullRequestName,
		AuthorID:          req.AuthorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: make([]string, 0, domain.MaxReviewersCount),
		CreatedAt:         &now,
	}

	reviewers := helpers.RandSelectReviewers(team.Members, req.AuthorID, domain.MaxReviewersCount)
	needMoreReviewers := len(reviewers) < domain.MaxReviewersCount

	if err := s.prRepo.CreatePullRequestWithReviewers(ctx, pr, reviewers, needMoreReviewers); err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	pr.NeedMoreReviewers = &needMoreReviewers

	helpers.UpdateReviewerLoadMetrics(ctx, s.prReviewersRepo, reviewers)

	logger.LogBusinessTransactionEnd(operation, time.Since(start), true, map[string]interface{}{
		"pr_id": req.PullRequestID,
	})
	logger.LogCriticalEvent("pull_request_created", map[string]interface{}{
		"pr_id": req.PullRequestID,
	})

	return pr, nil
}
