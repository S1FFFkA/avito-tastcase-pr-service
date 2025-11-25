package helpers

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
	"AVITOSAMPISHU/pkg/metrics"
	"context"
)

// UpdateReviewerLoadMetrics обновляет метрику распределения нагрузки для указанных ревьюверов
func UpdateReviewerLoadMetrics(
	ctx context.Context,
	prReviewersRepo repository.PrReviewersRepositoryInterface,
	reviewerIDs []string,
) {
	for _, reviewerID := range reviewerIDs {
		prs, err := prReviewersRepo.GetPRsByReviewer(ctx, reviewerID)
		if err != nil {
			continue
		}

		openPRCount := 0
		for _, pr := range prs {
			if pr.Status == domain.PRStatusOpen {
				openPRCount++
			}
		}

		metrics.ReviewerLoadDistribution.Observe(float64(openPRCount))
	}
}
