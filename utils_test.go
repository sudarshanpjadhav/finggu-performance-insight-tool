package main

import "testing"

func TestFingguFormatPrometheus_ContainsAllMetrics(t *testing.T) {
	metric := &FingguPerformanceMetric{
		CPUUsage:       55.5,
		MemoryUsage:    60.2,
		MemoryUsedMB:   1024,
		GoroutineCount: 42,
		HeapAllocMB:    12.3,
		NumGC:          7,
	}

	output := fingguFn_FormatPrometheus(metric)

	requiredSubstrings := []string{
		"finggu_cpu_usage_percent",
		"finggu_memory_usage_percent",
		"finggu_memory_used_mb",
		"finggu_goroutine_count",
		"finggu_heap_alloc_mb",
		"finggu_num_gc_total",
	}

	for _, substr := range requiredSubstrings {
		if !fingguFn_Contains(output, substr) {
			t.Errorf("expected prometheus output to contain %q", substr)
		}
	}
}

func fingguFn_Contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

func TestFingguFormatBytesToMB(t *testing.T) {
	got := fingguFn_FormatBytesToMB(1048576) // exactly 1 MB in bytes
	want := 1.0
	if got != want {
		t.Fatalf("expected %f, got %f", want, got)
	}
}
