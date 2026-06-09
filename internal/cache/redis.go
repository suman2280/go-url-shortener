package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"

	"github.com/suman2280/go-url-shortener/internal/domain"
)

var (
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RedisCache{client: rdb, ttl: ttl}
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Get(ctx context.Context, code string) (*domain.UrlMapping, error) {
	data, err := c.client.Get(ctx, code).Result()
	if err == redis.Nil {
		cacheMisses.Inc()
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	cacheHits.Inc()
	var m domain.UrlMapping
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *RedisCache) Set(ctx context.Context, url *domain.UrlMapping) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, url.ShortCode, data, c.ttl).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Client() *redis.Client {
	return c.client
}
