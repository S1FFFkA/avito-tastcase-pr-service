package helpers

import (
	"AVITOSAMPISHU/internal/domain"
	"math/rand"
	"time"
)

// RandSelectReviewers случайно выбирает ревьюверов из списка участников команды
// Исключает автора и неактивных пользователей
func RandSelectReviewers(members []domain.TeamMember, authorID string, maxCount int) []string {
	if maxCount <= 0 {
		return []string{}
	}

	candidates := make([]string, 0, len(members))
	for _, member := range members {
		if member.IsActive && member.UserID != authorID {
			candidates = append(candidates, member.UserID)
		}
	}

	if len(candidates) <= maxCount {
		return candidates
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	return candidates[:maxCount]
}
