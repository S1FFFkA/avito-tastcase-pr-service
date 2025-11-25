package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"time"

	"go.uber.org/zap"
)

func (s *PullRequestServiceImpl) ReassignReviewer(
	ctx context.Context,
	req *domain.ReassignReviewerReq,
) (*domain.PullRequest, string, error) {
	start := time.Now()
	operation := "ReassignReviewer"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"pr_id":        req.PullRequestID,
		"old_reviewer": req.OldUserID,
	})

	pr, err := s.prRepo.GetPullRequestByID(ctx, req.PullRequestID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": "PR already merged",
		})
		return nil, "", domain.ErrPRMerged
	}

	currentReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, "", err
	}

	if !helpers.ContainsReviewer(currentReviewers, req.OldUserID) {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": "reviewer not assigned",
		})
		return nil, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUserByID(ctx, req.OldUserID)
	if err != nil {
		return nil, "", err
	}

	team, err := s.teamRepo.GetTeamByName(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", err
	}
	onlyActiveCandidates := make([]domain.TeamMember, 0, len(team.Members))

	assignedSet := make(map[string]struct{}, len(currentReviewers))
	for _, reviewerID := range currentReviewers {
		assignedSet[reviewerID] = struct{}{}
	}

	for _, member := range team.Members {
		if !member.IsActive {
			continue
		}
		if member.UserID == pr.AuthorID {
			continue
		}
		if _, alreadyAssigned := assignedSet[member.UserID]; alreadyAssigned {
			continue
		}
		onlyActiveCandidates = append(onlyActiveCandidates, member)
	}

	logger.LogBusinessRule("select_replacement_reviewer", map[string]interface{}{
		"pr_id":            req.PullRequestID,
		"candidates_count": len(onlyActiveCandidates),
		"author_id":        pr.AuthorID,
	})

	candidates := helpers.RandSelectReviewers(onlyActiveCandidates, pr.AuthorID, 1)
	if len(candidates) == 0 {
		logger.LogBusinessRule("no_replacement_candidate", map[string]interface{}{
			"pr_id": req.PullRequestID,
		})
		if err := s.prRepo.SetNeedMoreReviewers(ctx, req.PullRequestID, true); err != nil {
			logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
				"pr_id": req.PullRequestID,
				"error": err.Error(),
			})
			return nil, "", err
		}
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id":  req.PullRequestID,
			"error":  "no replacement candidate",
			"reason": "no_candidate",
		})
		return nil, "", domain.ErrNoCandidate
	}
	newReviewerID := candidates[0]

	logger.Logger.Debug("reassigning reviewer",
		zap.String("pr_id", req.PullRequestID),
		zap.String("old_reviewer", req.OldUserID),
		zap.String("new_reviewer", newReviewerID),
	)

	if err := s.prReviewersRepo.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID, newReviewerID); err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, "", err
	}

	updatedReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"pr_id": req.PullRequestID,
			"error": err.Error(),
		})
		return nil, "", err
	}

	pr.AssignedReviewers = updatedReviewers

	affectedReviewers := make([]string, 0, 2)
	affectedReviewers = append(affectedReviewers, req.OldUserID)
	if newReviewerID != "" {
		affectedReviewers = append(affectedReviewers, newReviewerID)
	}
	helpers.UpdateReviewerLoadMetrics(ctx, s.prReviewersRepo, affectedReviewers)

	logger.LogBusinessTransactionEnd(operation, time.Since(start), true, map[string]interface{}{
		"pr_id": req.PullRequestID,
	})
	logger.LogCriticalEvent("reviewer_reassigned", map[string]interface{}{
		"pr_id": req.PullRequestID,
	})

	return pr, newReviewerID, nil
}
