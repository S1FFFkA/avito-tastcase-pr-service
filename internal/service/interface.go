package service

import (
	"AVITOSAMPISHU/internal/domain"
	"context"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, req *domain.SetIsActiveRequest) (*domain.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	DeactivateTeamMembers(ctx context.Context, req *domain.DeactivateTeamMembersReq) (*domain.DeactivateTeamMembersRes, error)
}

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, req *domain.CreatePullRequestReq) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, req *domain.MergePullRequestReq) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, req *domain.ReassignReviewerReq) (*domain.PullRequest, string, error)
}
