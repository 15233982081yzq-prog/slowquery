package analyzer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const slowQueryAnalyzerNameSpace = "slow_query_analyzer"

var (
	FilterProcessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: slowQueryAnalyzerNameSpace,
			Name:      "slow_query_filter_total",
			Help:      "This metric will provide the statistics about the slow query message was filtered total count",
		},
		[]string{"name", "status"},
	)

	StoreProcessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: slowQueryAnalyzerNameSpace,
			Name:      "slow_query_store_batch_put_total",
			Help:      "This metric will provide the statistics about the slow query store total count",
		},
		[]string{"name", "status"},
	)

	CacheProcessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: slowQueryAnalyzerNameSpace,
			Name:      "slow_query_cache_cap",
			Help:      "This metric will provide the statistics about the slow query cache cap total count",
		},
		[]string{"name", "status"},
	)
)

var (
	ServiceProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  slowQueryAnalyzerNameSpace,
			Name:       "service_process_duration_ms",
			Help:       "A summary of all service process duration_ms",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)

	StoreProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  slowQueryAnalyzerNameSpace,
			Name:       "store_process_duration_ms",
			Help:       "A summary of store process duration_ms",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)

	KafkaProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  slowQueryAnalyzerNameSpace,
			Name:       "slow_query_kafka_consume_duration_ms",
			Help:       "This metric will provide the statistics about the slow query kafka consume message duration_ms",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)

	FingerFilterExistProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  slowQueryAnalyzerNameSpace,
			Name:       "slow_query_analyzer_finger_filter_exist_duration_ms",
			Help:       "This metric will provide the statistics about the slow query finger filter exist cost",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)

	FingerFilterFreshProcessSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  slowQueryAnalyzerNameSpace,
			Name:       "slow_query_analyzer_finger_filter_fresh_duration_ms",
			Help:       "This metric will provide the statistics about the slow query finger filter fresh cost",
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.90: 0.01, 0.95: 0.01, 0.99: 0.001}, //nolint:gomnd
		},
		[]string{"name", "status"},
	)
)
