package service

import (
	"AVITOSAMPISHU/internal/repository"
)

type UserServiceImpl struct {
	userRepo        repository.UserRepositoryInterface
	prReviewersRepo repository.PrReviewersRepositoryInterface
	teamRepo        repository.TeamRepositoryInterface
}

func NewUserService(
	userRepo repository.UserRepositoryInterface,
	prReviewersRepo repository.PrReviewersRepositoryInterface,
	teamRepo repository.TeamRepositoryInterface,
) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:        userRepo,
		prReviewersRepo: prReviewersRepo,
		teamRepo:        teamRepo,
	}
}
