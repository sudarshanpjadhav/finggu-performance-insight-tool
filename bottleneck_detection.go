package main

import "fmt"

// FingguDetectBottleneck evaluates a metric snapshot against configured
// thresholds and returns a structured result explaining exactly which
// signals tripped, instead of a black-box true/false.
func FingguDetectBottleneck(metric *FingguPerformanceMetric, cfg *FingguConfig) FingguBottleneckResult {
	reasons := make([]string, 0, 3)

	if metric.CPUUsage >= cfg.CPUThresholdPercent {
		reasons = append(reasons, fmt.Sprintf(
			"CPU usage %.1f%% exceeds threshold %.1f%%", metric.CPUUsage, cfg.CPUThresholdPercent))
	}

	if metric.MemoryUsage >= cfg.MemThresholdPercent {
		reasons = append(reasons, fmt.Sprintf(
			"Memory usage %.1f%% exceeds threshold %.1f%%", metric.MemoryUsage, cfg.MemThresholdPercent))
	}

	if metric.GoroutineCount >= cfg.GoroutineThreshold {
		reasons = append(reasons, fmt.Sprintf(
			"Goroutine count %d exceeds threshold %d (possible leak)", metric.GoroutineCount, cfg.GoroutineThreshold))
	}

	return FingguBottleneckResult{
		Detected: len(reasons) > 0,
		Reasons:  reasons,
		Metric:   *metric,
	}
}
