package platform

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const SlowQueryPromNamespace = "slow_query_openapi"

var (
	ServiceProcessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: SlowQueryPromNamespace,
			Name:      "slow_query_service_total",
			Help:      "This metric will provide the statistics about the slow query service count",
		},
		[]string{"name", "status"},
	)

	StoreProcessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: SlowQueryPromNamespace,
			Name:      "slow_query_store_total",
			Help:      "This metric will provide the statistics about the slow query store count",
		},
		[]string{"name", "status"},
	)
)

var (
	ServiceProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  SlowQueryPromNamespace,
			Name:       "service_process_duration_ms",
			Help:       "A summary of all service process duration_ms",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)

	StoreProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  SlowQueryPromNamespace,
			Name:       "store_process_duration_ms",
			Help:       "A summary of store process duration_ms",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)
)
