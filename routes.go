package main

import (
	"github.com/gin-gonic/gin"
)

// FingguSetupRoutes registers every HTTP endpoint the service exposes.
func FingguSetupRoutes(router *gin.Engine, api *FingguAPI) {
	router.GET("/health", api.FingguHandleHealth)
	router.GET("/metrics", api.FingguHandlePrometheusMetrics)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/performance", api.FingguHandleCurrentMetrics)
		v1.GET("/performance/history", api.FingguHandleHistory)
		v1.GET("/performance/bottleneck", api.FingguHandleBottleneckCheck)
		v1.GET("/config", api.FingguHandleConfig)
	}

	// Kept for backwards compatibility with the original v0 route.
	router.GET("/performance", api.FingguHandleCurrentMetrics)
}
