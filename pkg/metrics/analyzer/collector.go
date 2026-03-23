package analyzer

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

func CollectStoreBatchCounter(name, status string, count int) {
	StoreProcessCounter.WithLabelValues(name, status).Add(float64(count))
}

func CollectFilterCounter(name, status string, count int) {
	CacheProcessCounter.WithLabelValues(name, status).Add(float64(count))
}

func CollectFilterCap(name, status string, count int) {
	CacheProcessCounter.WithLabelValues(name, status).Add(float64(count))
}

func CollectFilterExistMetrics(name, status string, cost time.Duration) {
	FingerFilterExistProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
}

func CollectFilterFreshMetrics(name, status string, cost time.Duration) {
	FingerFilterFreshProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
}

func CollectMessageFilteredCounter(name, status string, count int) {
	FilterProcessCounter.WithLabelValues(name, status).Add(float64(count))
}

func CollectKafkaMetrics(name, status string, cost time.Duration) {
	KafkaProcessSummary.WithLabelValues(name, status).Observe(float64(cost.Milliseconds()))
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

func GetExistStatus(exist bool) (status string) {
	if exist {
		return "true"
	}
	return "false"
}
