package main

import (
	"encoding/json"
	"time"
)

// FingguPerformanceMetric represents a single snapshot of system performance
// collected by the FingguCollector. It's the core data unit stored in Redis
// and returned by every API endpoint.
type FingguPerformanceMetric struct {
	ID             string    `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	CPUUsage       float64   `json:"cpu_usage_percent"`
	MemoryUsage    float64   `json:"memory_usage_percent"`
	MemoryUsedMB   float64   `json:"memory_used_mb"`
	MemoryTotalMB  float64   `json:"memory_total_mb"`
	GoroutineCount int       `json:"goroutine_count"`
	NumGC          uint32    `json:"num_gc"`
	HeapAllocMB    float64   `json:"heap_alloc_mb"`
}

// FingguBottleneckResult carries the outcome of a threshold evaluation
// against a metric snapshot, including *why* it tripped so alerts are
// actionable instead of generic.
type FingguBottleneckResult struct {
	Detected bool                    `json:"detected"`
	Reasons  []string                `json:"reasons"`
	Metric   FingguPerformanceMetric `json:"metric"`
}

// ToJSON serializes the metric for storage or transport.
func (metric *FingguPerformanceMetric) ToJSON() (string, error) {
	data, err := json.Marshal(metric)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FingguMetricFromJSON deserializes a metric previously stored via ToJSON.
func FingguMetricFromJSON(data string) (*FingguPerformanceMetric, error) {
	var metric FingguPerformanceMetric
	if err := json.Unmarshal([]byte(data), &metric); err != nil {
		return nil, err
	}
	return &metric, nil
}
