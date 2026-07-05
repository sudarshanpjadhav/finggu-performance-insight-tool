package main

import "fmt"

// fingguFn_FormatPrometheus renders a metric snapshot as Prometheus text
// exposition format (https://prometheus.io/docs/instrumenting/exposition_formats/).
func fingguFn_FormatPrometheus(metric *FingguPerformanceMetric) string {
	return fmt.Sprintf(
		"# HELP finggu_cpu_usage_percent Current CPU usage percentage\n"+
			"# TYPE finggu_cpu_usage_percent gauge\n"+
			"finggu_cpu_usage_percent %.2f\n"+
			"# HELP finggu_memory_usage_percent Current memory usage percentage\n"+
			"# TYPE finggu_memory_usage_percent gauge\n"+
			"finggu_memory_usage_percent %.2f\n"+
			"# HELP finggu_memory_used_mb Memory used in megabytes\n"+
			"# TYPE finggu_memory_used_mb gauge\n"+
			"finggu_memory_used_mb %.2f\n"+
			"# HELP finggu_goroutine_count Number of active goroutines\n"+
			"# TYPE finggu_goroutine_count gauge\n"+
			"finggu_goroutine_count %d\n"+
			"# HELP finggu_heap_alloc_mb Go heap allocation in megabytes\n"+
			"# TYPE finggu_heap_alloc_mb gauge\n"+
			"finggu_heap_alloc_mb %.2f\n"+
			"# HELP finggu_num_gc_total Total number of completed GC cycles\n"+
			"# TYPE finggu_num_gc_total counter\n"+
			"finggu_num_gc_total %d\n",
		metric.CPUUsage, metric.MemoryUsage, metric.MemoryUsedMB,
		metric.GoroutineCount, metric.HeapAllocMB, metric.NumGC,
	)
}

// fingguFn_FormatBytesToMB is a small helper kept for readability at call
// sites that only have a raw byte count.
func fingguFn_FormatBytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}
