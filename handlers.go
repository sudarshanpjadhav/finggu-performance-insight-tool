package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FingguAPI bundles the shared dependencies every HTTP handler needs, so
// routes.go can register bound methods instead of relying on globals.
type FingguAPI struct {
	redis *FingguRedisClient
	cfg   *FingguConfig
}

// FingguNewAPI constructs the handler group.
func FingguNewAPI(redisClient *FingguRedisClient, cfg *FingguConfig) *FingguAPI {
	return &FingguAPI{redis: redisClient, cfg: cfg}
}

// FingguHandleCurrentMetrics returns the most recently collected snapshot.
func (api *FingguAPI) FingguHandleCurrentMetrics(c *gin.Context) {
	metric, err := api.redis.GetLatestMetric()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metric)
}

// FingguHandleHistory returns recent metrics, newest first, with an
// optional ?limit= query param (defaults to 50).
func (api *FingguAPI) FingguHandleHistory(c *gin.Context) {
	limit := 50
	if q := c.Query("limit"); q != "" {
		if parsed, err := strconv.Atoi(q); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := api.redis.GetHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": len(history), "metrics": history})
}

// FingguHandleBottleneckCheck runs detection against the latest snapshot
// on demand, useful for CI pipelines polling a single pass/fail signal.
func (api *FingguAPI) FingguHandleBottleneckCheck(c *gin.Context) {
	metric, err := api.redis.GetLatestMetric()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	result := FingguDetectBottleneck(metric, api.cfg)
	c.JSON(http.StatusOK, result)
}

// FingguHandleConfig exposes the active (non-secret) thresholds so
// dashboards and CI scripts can display what's currently enforced.
func (api *FingguAPI) FingguHandleConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cpu_threshold_percent":  api.cfg.CPUThresholdPercent,
		"mem_threshold_percent":  api.cfg.MemThresholdPercent,
		"goroutine_threshold":    api.cfg.GoroutineThreshold,
		"poll_interval_seconds":  api.cfg.PollInterval.Seconds(),
		"alert_cooldown_seconds": api.cfg.AlertCooldown.Seconds(),
		"environment":            api.cfg.Environment,
	})
}

// FingguHandleHealth is a liveness/readiness probe for k8s, Docker
// healthchecks, and uptime monitors. It confirms Redis is reachable.
func (api *FingguAPI) FingguHandleHealth(c *gin.Context) {
	if err := api.redis.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "redis": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// FingguHandlePrometheusMetrics exposes current metrics in Prometheus text
// exposition format so the tool drops straight into an existing scrape
// config without needing a separate exporter.
func (api *FingguAPI) FingguHandlePrometheusMetrics(c *gin.Context) {
	metric, err := api.redis.GetLatestMetric()
	if err != nil {
		c.String(http.StatusNotFound, "# no metrics recorded yet\n")
		return
	}
	c.String(http.StatusOK, fingguFn_FormatPrometheus(metric))
}
