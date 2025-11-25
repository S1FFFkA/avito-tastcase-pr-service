package service

import (
	"AVITOSAMPISHU/internal/repository"
)

type PullRequestServiceImpl struct {
	prRepo          repository.PullRequestRepositoryInterface
	prReviewersRepo repository.PrReviewersRepositoryInterface
	userRepo        repository.UserRepositoryInterface
	teamRepo        repository.TeamRepositoryInterface
}

func NewPullRequestService(
	prRepo repository.PullRequestRepositoryInterface,
	prReviewersRepo repository.PrReviewersRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	teamRepo repository.TeamRepositoryInterface,
) *PullRequestServiceImpl {
	return &PullRequestServiceImpl{
		prRepo:          prRepo,
		prReviewersRepo: prReviewersRepo,
		userRepo:        userRepo,
		teamRepo:        teamRepo,
	}
}
