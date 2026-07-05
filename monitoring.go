package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// FingguCollector owns the background polling loop: it gathers real host
// metrics, persists them, runs bottleneck detection, and fires alerts.
type FingguCollector struct {
	cfg          *FingguConfig
	redis        *FingguRedisClient
	alertManager *FingguAlertManager
	stopCh       chan struct{}
}

// FingguNewCollector wires up a collector with its dependencies.
func FingguNewCollector(cfg *FingguConfig, redisClient *FingguRedisClient, alertManager *FingguAlertManager) *FingguCollector {
	return &FingguCollector{
		cfg:          cfg,
		redis:        redisClient,
		alertManager: alertManager,
		stopCh:       make(chan struct{}),
	}
}

// FingguCollectSnapshot gathers a single real metric sample from the host
// process/OS. CPU percent is measured over a short blocking sample window.
func FingguCollectSnapshot() (*FingguPerformanceMetric, error) {
	cpuPercents, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		return nil, fmt.Errorf("finggu: cpu sample failed: %w", err)
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("finggu: memory sample failed: %w", err)
	}

	var goRuntimeStats runtime.MemStats
	runtime.ReadMemStats(&goRuntimeStats)

	metric := &FingguPerformanceMetric{
		ID:             fmt.Sprintf("finggu-%d", time.Now().UnixNano()),
		Timestamp:      time.Now().UTC(),
		CPUUsage:       cpuUsage,
		MemoryUsage:    vmStat.UsedPercent,
		MemoryUsedMB:   float64(vmStat.Used) / 1024 / 1024,
		MemoryTotalMB:  float64(vmStat.Total) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		NumGC:          goRuntimeStats.NumGC,
		HeapAllocMB:    float64(goRuntimeStats.HeapAlloc) / 1024 / 1024,
	}
	return metric, nil
}

// Start begins the polling loop on a ticker; it runs until Stop is called.
func (collector *FingguCollector) Start() {
	ticker := time.NewTicker(collector.cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("finggu: monitoring started (interval=%s)", collector.cfg.PollInterval)

	for {
		select {
		case <-ticker.C:
			collector.fingguRunCycle()
		case <-collector.stopCh:
			log.Println("finggu: monitoring stopped")
			return
		}
	}
}

// Stop signals the polling loop to exit gracefully.
func (collector *FingguCollector) Stop() {
	close(collector.stopCh)
}

func (collector *FingguCollector) fingguRunCycle() {
	metric, err := FingguCollectSnapshot()
	if err != nil {
		log.Printf("finggu: snapshot collection error: %v", err)
		return
	}

	if err := collector.redis.SaveMetric(metric); err != nil {
		log.Printf("finggu: failed to persist metric: %v", err)
	}

	result := FingguDetectBottleneck(metric, collector.cfg)
	if result.Detected {
		collector.alertManager.FingguDispatch(result)
	}
}
