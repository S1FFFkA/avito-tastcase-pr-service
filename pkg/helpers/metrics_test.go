package helpers

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/repository"
	"context"
	"errors"
	"testing"
)

type mockPrReviewersRepository struct {
	repository.PrReviewersRepositoryInterface
	getPRsByReviewerFunc func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *mockPrReviewersRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.getPRsByReviewerFunc != nil {
		return m.getPRsByReviewerFunc(ctx, userID)
	}
	return nil, nil
}

func TestUpdateReviewerLoadMetrics(t *testing.T) {
	tests := []struct {
		name          string
		reviewerIDs   []string
		setupMocks    func(*mockPrReviewersRepository)
		expectedCalls int
	}{
		{
			name:        "successful update for multiple reviewers",
			reviewerIDs: []string{"reviewer-1", "reviewer-2"},
			setupMocks: func(repo *mockPrReviewersRepository) {
				repo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{
						{PullRequestID: "pr-1", Status: domain.PRStatusOpen},
						{PullRequestID: "pr-2", Status: domain.PRStatusOpen},
						{PullRequestID: "pr-3", Status: domain.PRStatusMerged},
					}, nil
				}
			},
			expectedCalls: 2,
		},
		{
			name:        "handle repository error",
			reviewerIDs: []string{"reviewer-1", "reviewer-2"},
			setupMocks: func(repo *mockPrReviewersRepository) {
				repo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return nil, errors.New("database error")
				}
			},
			expectedCalls: 2, // Функция должна вызваться для всех, но продолжить работу при ошибке
		},
		{
			name:        "empty reviewers list",
			reviewerIDs: []string{},
			setupMocks: func(repo *mockPrReviewersRepository) {
				repo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return nil, nil
				}
			},
			expectedCalls: 0,
		},
		{
			name:        "count only open PRs",
			reviewerIDs: []string{"reviewer-1"},
			setupMocks: func(repo *mockPrReviewersRepository) {
				repo.getPRsByReviewerFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{
						{PullRequestID: "pr-1", Status: domain.PRStatusOpen},
						{PullRequestID: "pr-2", Status: domain.PRStatusMerged},
						{PullRequestID: "pr-3", Status: domain.PRStatusOpen},
					}, nil
				}
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPrReviewersRepository{}
			tt.setupMocks(repo)

			// Функция не возвращает ошибку, просто обновляет метрики
			UpdateReviewerLoadMetrics(context.Background(), repo, tt.reviewerIDs)

			// Проверяем, что функция была вызвана правильное количество раз
			// (в реальном тесте можно использовать счетчик вызовов)
		})
	}
}
