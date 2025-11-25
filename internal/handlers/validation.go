package handlers

import (
	"fmt"

	"AVITOSAMPISHU/internal/domain"
)

func validateTeam(team *domain.Team) error {
	if team.TeamName == "" {
		return fmt.Errorf("%w: team_name is required", domain.ErrInvalidRequest)
	}
	if len(team.Members) == 0 {
		return fmt.Errorf("%w: team must have at least one member", domain.ErrInvalidRequest)
	}
	for i, member := range team.Members {
		if member.UserID == "" {
			return fmt.Errorf("%w: member[%d].user_id is required", domain.ErrInvalidRequest, i)
		}
		if member.Username == "" {
			return fmt.Errorf("%w: member[%d].username is required", domain.ErrInvalidRequest, i)
		}
	}
	return nil
}

func validateSetIsActiveRequest(req *domain.SetIsActiveRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("%w: user_id is required", domain.ErrInvalidRequest)
	}
	return nil
}

func validateDeactivateTeamMembersReq(req *domain.DeactivateTeamMembersReq) error {
	if req.TeamName == "" {
		return fmt.Errorf("%w: team_name is required", domain.ErrInvalidRequest)
	}
	if len(req.UserIDs) > 0 {
		for i, userID := range req.UserIDs {
			if userID == "" {
				return fmt.Errorf("%w: user_ids[%d] cannot be empty", domain.ErrInvalidRequest, i)
			}
		}
	}
	return nil
}

func validateCreatePullRequestReq(req *domain.CreatePullRequestReq) error {
	if req.PullRequestID == "" {
		return fmt.Errorf("%w: pull_request_id is required", domain.ErrInvalidRequest)
	}
	if req.PullRequestName == "" {
		return fmt.Errorf("%w: pull_request_name is required", domain.ErrInvalidRequest)
	}
	if req.AuthorID == "" {
		return fmt.Errorf("%w: author_id is required", domain.ErrInvalidRequest)
	}
	return nil
}

func validateMergePullRequestReq(req *domain.MergePullRequestReq) error {
	if req.PullRequestID == "" {
		return fmt.Errorf("%w: pull_request_id is required", domain.ErrInvalidRequest)
	}
	return nil
}

func validateReassignReviewerReq(req *domain.ReassignReviewerReq) error {
	if req.PullRequestID == "" {
		return fmt.Errorf("%w: pull_request_id is required", domain.ErrInvalidRequest)
	}
	if req.OldUserID == "" {
		return fmt.Errorf("%w: old_user_id is required", domain.ErrInvalidRequest)
	}
	return nil
}
