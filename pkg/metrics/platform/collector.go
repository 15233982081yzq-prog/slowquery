package platform

import (
	errStore "smart-slowquery/internal/util/errors"

	"time"
)

func CollectServiceMetrics(name, status string, cost time.Duration) {
	ServiceProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
}

func CollectStoreMetrics(name, status string, cost time.Duration) {
	StoreProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
}

func CollectExplainMetrics(name, status string, cost time.Duration) {
	ExplainProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
}

func GetStatus(err error) (status string) {
	switch err {
	case nil:
		status = "success"
	case errStore.NotFoundError:
		status = "success"
	default:
		status = "failed"
	}
	return
}
