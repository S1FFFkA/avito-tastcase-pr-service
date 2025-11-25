package service

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"time"
)

// SetIsActive обновляет флаг активности пользователя через репозиторий пользователей.
func (s *UserServiceImpl) SetIsActive(ctx context.Context, req *domain.SetIsActiveRequest) (*domain.User, error) {
	start := time.Now()
	operation := "SetIsActive"

	logger.LogBusinessTransactionStart(operation, map[string]interface{}{
		"user_id":   req.UserID,
		"is_active": req.IsActive,
	})

	if err := s.userRepo.SetUserIsActive(ctx, req.UserID, req.IsActive); err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		logger.LogBusinessTransactionEnd(operation, time.Since(start), false, map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, err
	}

	duration := time.Since(start)
	logger.LogBusinessTransactionEnd(operation, duration, true, map[string]interface{}{
		"user_id":   user.UserID,
		"is_active": user.IsActive,
	})

	return user, nil
}
