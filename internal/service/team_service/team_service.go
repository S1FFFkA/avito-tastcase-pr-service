package service

import (
	"AVITOSAMPISHU/internal/repository"
)

type TeamServiceImpl struct {
	teamRepo repository.TeamRepositoryInterface
	userRepo repository.UserRepositoryInterface
}

func NewTeamService(
	teamRepo repository.TeamRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
) *TeamServiceImpl {
	return &TeamServiceImpl{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}
