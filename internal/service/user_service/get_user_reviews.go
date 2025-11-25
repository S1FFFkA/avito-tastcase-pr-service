package service

import (
	"AVITOSAMPISHU/internal/domain"
	"context"
)

// GetUserReviews возвращает список PR, где пользователь назначен ревьювером.
func (s *UserServiceImpl) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	prs, err := s.prReviewersRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
