package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// ReviewerLoadDistribution отслеживает распределение нагрузки между ревьюверами
// Показывает, сколько PR одновременно назначено каждому ревьюверу
// Бакеты: 0, 1, 2, 3, 5, 10, 15, 20, 30, 50
var ReviewerLoadDistribution = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name: "reviewer_load_distribution",
	Help: "Distribution of PR assignments per reviewer (shows workload balance)",
	Buckets: func() []float64 {
		buckets := make([]float64, 0, 10)
		buckets = append(buckets, 0, 1, 2, 3, 5, 10, 15, 20, 30, 50)
		return buckets
	}(),
})
