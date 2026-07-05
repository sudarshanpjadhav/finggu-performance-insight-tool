package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// FingguRedisClient wraps the go-redis client with the specific list-backed
// time series operations this tool needs: push a new sample, trim old ones,
// and read back the most recent N.
type FingguRedisClient struct {
	Client       *redis.Client
	ctx          context.Context
	seriesKey    string
	maxRetention int
}

const fingguMetricSeriesKey = "finggu:metrics:series"
const fingguLatestMetricKey = "finggu:metrics:latest"

// FingguNewRedisClient builds a connected client wrapper from config.
func FingguNewRedisClient(cfg *FingguConfig) *FingguRedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &FingguRedisClient{
		Client:       rdb,
		ctx:          context.Background(),
		seriesKey:    fingguMetricSeriesKey,
		maxRetention: cfg.MetricRetention,
	}
}

// Ping verifies connectivity, used by the /health endpoint and at startup.
func (client *FingguRedisClient) Ping() error {
	timeoutCtx, cancel := context.WithTimeout(client.ctx, 2*time.Second)
	defer cancel()
	return client.Client.Ping(timeoutCtx).Err()
}

// SaveMetric persists a metric snapshot into a capped Redis list (acting as
// a lightweight time series) and updates the "latest" pointer key.
func (client *FingguRedisClient) SaveMetric(metric *FingguPerformanceMetric) error {
	payload, err := metric.ToJSON()
	if err != nil {
		return fmt.Errorf("finggu: failed to serialize metric: %w", err)
	}

	pipe := client.Client.TxPipeline()
	pipe.LPush(client.ctx, client.seriesKey, payload)
	pipe.LTrim(client.ctx, client.seriesKey, 0, int64(client.maxRetention-1))
	pipe.Set(client.ctx, fingguLatestMetricKey, payload, 0)
	_, err = pipe.Exec(client.ctx)
	if err != nil {
		return fmt.Errorf("finggu: failed to save metric to redis: %w", err)
	}
	return nil
}

// GetLatestMetric returns the most recently recorded snapshot.
func (client *FingguRedisClient) GetLatestMetric() (*FingguPerformanceMetric, error) {
	val, err := client.Client.Get(client.ctx, fingguLatestMetricKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("finggu: no metrics recorded yet")
	} else if err != nil {
		return nil, err
	}
	return FingguMetricFromJSON(val)
}

// GetHistory returns up to `limit` of the most recent metrics, newest first.
func (client *FingguRedisClient) GetHistory(limit int) ([]*FingguPerformanceMetric, error) {
	if limit <= 0 || limit > client.maxRetention {
		limit = client.maxRetention
	}
	raw, err := client.Client.LRange(client.ctx, client.seriesKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	metrics := make([]*FingguPerformanceMetric, 0, len(raw))
	for _, item := range raw {
		m, err := FingguMetricFromJSON(item)
		if err != nil {
			continue // skip malformed/legacy entries instead of failing the whole request
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}
