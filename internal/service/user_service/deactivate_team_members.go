package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/helpers"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"fmt"
	"time"
)

// DeactivateTeamMembers деактивирует участников команды и, при необходимости,
// подготавливает план переназначения ревьюверов для их открытых PR.
func (s *UserServiceImpl) DeactivateTeamMembers(
	ctx context.Context,
	req *domain.DeactivateTeamMembersReq,
) (*domain.DeactivateTeamMembersRes, error) {
	start := time.Now()
	operation := "DeactivateTeamMembers"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"team_name":      req.TeamName,
		"users_count":    len(req.UserIDs),
		"deactivate_all": len(req.UserIDs) == 0,
	})

	team, err := s.teamRepo.GetTeamByName(ctx, req.TeamName)
	// Ошибки репозитория уже доменные, поэтому просто пробрасываем их выше.
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	// Проверка: нельзя деактивировать всех участников команды без явного указания UserIDs
	if len(req.UserIDs) == 0 {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     "cannot deactivate all team members without explicit user IDs",
			"reason":    "empty_user_ids",
		})
		return nil, fmt.Errorf("%w: cannot deactivate all team members without explicit user IDs", domain.ErrInvalidRequest)
	}

	// Проверка: нельзя деактивировать всех участников команды
	if len(req.UserIDs) == len(team.Members) {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     "cannot deactivate all team members, team would be left without reviewers",
			"reason":    "deactivate_all_members",
		})
		return nil, fmt.Errorf("%w: cannot deactivate all team members, team would be left without reviewers", domain.ErrInvalidRequest)
	}

	memberIndex := make(map[string]domain.TeamMember, len(team.Members))
	for _, member := range team.Members {
		memberIndex[member.UserID] = member
	}

	for _, userID := range req.UserIDs {
		if _, ok := memberIndex[userID]; !ok {
			logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
				"team_name": req.TeamName,
				"user_id":   userID,
				"error":     "user is not a member of team",
				"reason":    "invalid_member",
			})
			return nil, fmt.Errorf("%w: user %s is not a member of team %s", domain.ErrInvalidRequest, userID, req.TeamName)
		}
	}

	prMap, err := s.getOpenPRsForUsers(ctx, req.UserIDs)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	reassignments, err := s.buildReassignmentsPlan(ctx, prMap, req.UserIDs, team)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	deactivatedUserIDs, err := s.teamRepo.DeactivateTeamMembers(ctx, req.TeamName, req.UserIDs, reassignments)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"team_name": req.TeamName,
			"error":     err.Error(),
		})
		return nil, err
	}

	affectedReviewers := make([]string, 0, len(deactivatedUserIDs)+len(reassignments))
	affectedReviewers = append(affectedReviewers, deactivatedUserIDs...)
	for _, reassignment := range reassignments {
		if reassignment.NewReviewerID != "" {
			affectedReviewers = append(affectedReviewers, reassignment.NewReviewerID)
		}
	}
	helpers.UpdateReviewerLoadMetrics(ctx, s.prReviewersRepo, affectedReviewers)

	logger.LogBusinessTransactionEnd(operation, time.Since(start), true, map[string]interface{}{
		"team_name": req.TeamName,
	})
	logger.LogCriticalEvent("team_members_deactivated", map[string]interface{}{
		"team_name": req.TeamName,
	})

	return &domain.DeactivateTeamMembersRes{
		DeactivatedUserIDs: deactivatedUserIDs,
		Reassignments:      reassignments,
	}, nil
}

func (s *UserServiceImpl) getOpenPRsForUsers(
	ctx context.Context,
	userIDs []string,
) (map[string]domain.PullRequestShort, error) {
	prMap := make(map[string]domain.PullRequestShort, len(userIDs))

	for _, userID := range userIDs {
		prs, err := s.prReviewersRepo.GetPRsByReviewer(ctx, userID)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			if pr.Status == domain.PRStatusOpen {
				prMap[pr.PullRequestID] = pr
			}
		}
	}

	return prMap, nil
}

func (s *UserServiceImpl) buildReassignmentsPlan(
	ctx context.Context,
	prMap map[string]domain.PullRequestShort,
	usersToDeactivate []string,
	team *domain.Team,
) ([]domain.ReviewerReassignment, error) {
	reassignments := make([]domain.ReviewerReassignment, 0, len(prMap))

	usersToDeactivateSet := make(map[string]struct{}, len(usersToDeactivate))
	for _, userID := range usersToDeactivate {
		usersToDeactivateSet[userID] = struct{}{}
	}

	availableMembers := make([]domain.TeamMember, 0, len(team.Members))
	for _, member := range team.Members {
		if _, marked := usersToDeactivateSet[member.UserID]; marked {
			continue
		}
		if !member.IsActive {
			continue
		}
		availableMembers = append(availableMembers, member)
	}

	for _, pr := range prMap {
		currentReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, err
		}

		reviewersToReplace := make([]string, 0, len(currentReviewers))
		alreadyAssigned := make(map[string]struct{}, len(currentReviewers))

		for _, reviewerID := range currentReviewers {
			alreadyAssigned[reviewerID] = struct{}{}
			if _, needReplace := usersToDeactivateSet[reviewerID]; needReplace {
				reviewersToReplace = append(reviewersToReplace, reviewerID)
			}
		}

		if len(reviewersToReplace) == 0 {
			continue
		}

		candidates := helpers.RandSelectReviewers(availableMembers, pr.AuthorID, len(reviewersToReplace))

		availableCandidates := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			if _, exists := alreadyAssigned[candidate]; !exists {
				availableCandidates = append(availableCandidates, candidate)
			}
		}

		candidateIndex := 0
		addedCount := 0

		for _, reviewerID := range reviewersToReplace {
			var newReviewerID string

			if candidateIndex < len(availableCandidates) {
				newReviewerID = availableCandidates[candidateIndex]
				alreadyAssigned[newReviewerID] = struct{}{}
				candidateIndex++
				addedCount++
			}

			reassignments = append(reassignments, domain.ReviewerReassignment{
				PrID:          pr.PullRequestID,
				OldReviewerID: reviewerID,
				NewReviewerID: newReviewerID,
			})
		}

		finalReviewerCount := len(currentReviewers) - len(reviewersToReplace) + addedCount
		if finalReviewerCount == 0 {
			return nil, fmt.Errorf("%w: PR %s would be left without reviewers", domain.ErrNoCandidate, pr.PullRequestID)
		}
	}

	return reassignments, nil
}
