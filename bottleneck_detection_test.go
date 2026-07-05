package main

import "testing"

func fingguFn_TestConfig() *FingguConfig {
	return &FingguConfig{
		CPUThresholdPercent: 80.0,
		MemThresholdPercent: 85.0,
		GoroutineThreshold:  5000,
	}
}

func TestFingguDetectBottleneck_NoIssues(t *testing.T) {
	cfg := fingguFn_TestConfig()
	metric := &FingguPerformanceMetric{CPUUsage: 20, MemoryUsage: 40, GoroutineCount: 10}

	result := FingguDetectBottleneck(metric, cfg)

	if result.Detected {
		t.Fatalf("expected no bottleneck, got reasons: %v", result.Reasons)
	}
}

func TestFingguDetectBottleneck_HighCPU(t *testing.T) {
	cfg := fingguFn_TestConfig()
	metric := &FingguPerformanceMetric{CPUUsage: 95, MemoryUsage: 40, GoroutineCount: 10}

	result := FingguDetectBottleneck(metric, cfg)

	if !result.Detected {
		t.Fatal("expected bottleneck to be detected for high CPU")
	}
	if len(result.Reasons) != 1 {
		t.Fatalf("expected exactly 1 reason, got %d: %v", len(result.Reasons), result.Reasons)
	}
}

func TestFingguDetectBottleneck_MultipleIssues(t *testing.T) {
	cfg := fingguFn_TestConfig()
	metric := &FingguPerformanceMetric{CPUUsage: 99, MemoryUsage: 99, GoroutineCount: 9999}

	result := FingguDetectBottleneck(metric, cfg)

	if !result.Detected {
		t.Fatal("expected bottleneck to be detected")
	}
	if len(result.Reasons) != 3 {
		t.Fatalf("expected 3 reasons, got %d: %v", len(result.Reasons), result.Reasons)
	}
}

func TestFingguDetectBottleneck_BoundaryIsInclusive(t *testing.T) {
	cfg := fingguFn_TestConfig()
	metric := &FingguPerformanceMetric{CPUUsage: 80.0, MemoryUsage: 40, GoroutineCount: 10}

	result := FingguDetectBottleneck(metric, cfg)

	if !result.Detected {
		t.Fatal("expected threshold value itself to trigger detection")
	}
}
