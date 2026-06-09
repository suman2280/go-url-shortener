package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/suman2280/go-url-shortener/internal/domain"
	"github.com/suman2280/go-url-shortener/internal/repository"
	"github.com/suman2280/go-url-shortener/internal/service"
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func skipIfNoService(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("skipping integration test")
	}
}

func TestPostgresIntegration_CreateAndGet(t *testing.T) {
	skipIfNoService(t)

	dsn := getEnvOrDefault("DATABASE_URL", "host=localhost user=urlshort password=urlshort dbname=urlshortener port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	db.AutoMigrate(&domain.UrlMapping{})
	db.Where("1 = 1").Delete(&domain.UrlMapping{})

	repo := repository.NewPostgresUrlRepository(db)

	now := time.Now().Add(1 * time.Hour)
	mapping := &domain.UrlMapping{
		ShortCode: "test01",
		LongUrl:   "https://example.com",
		ExpiresAt: &now,
	}

	if err := repo.Create(mapping); err != nil {
		t.Fatalf("failed to create mapping: %v", err)
	}

	got, err := repo.GetByCode("test01")
	if err != nil {
		t.Fatalf("failed to get mapping: %v", err)
	}
	if got.LongUrl != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", got.LongUrl)
	}
}

func TestPostgresIntegration_CodeExists(t *testing.T) {
	skipIfNoService(t)

	dsn := getEnvOrDefault("DATABASE_URL", "host=localhost user=urlshort password=urlshort dbname=urlshortener port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	db.AutoMigrate(&domain.UrlMapping{})
	db.Where("1 = 1").Delete(&domain.UrlMapping{})

	repo := repository.NewPostgresUrlRepository(db)

	now := time.Now().Add(1 * time.Hour)
	repo.Create(&domain.UrlMapping{
		ShortCode: "exists1",
		LongUrl:   "https://example.com",
		ExpiresAt: &now,
	})

	exists, err := repo.CodeExists("exists1")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("expected code to exist")
	}

	exists, err = repo.CodeExists("nonexist")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("expected code to not exist")
	}
}

func TestRedisIntegration(t *testing.T) {
	skipIfNoService(t)

	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}

	if err := rdb.Set(ctx, "testkey", "testvalue", 1*time.Hour).Err(); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	val, err := rdb.Get(ctx, "testkey").Result()
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if val != "testvalue" {
		t.Errorf("expected testvalue, got %s", val)
	}

	rdb.Del(ctx, "testkey")
}

func TestServiceIntegration_GetByCode(t *testing.T) {
	skipIfNoService(t)

	dsn := getEnvOrDefault("DATABASE_URL", "host=localhost user=urlshort password=urlshort dbname=urlshortener port=5432 sslmode=disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	db.AutoMigrate(&domain.UrlMapping{})
	db.Where("1 = 1").Delete(&domain.UrlMapping{})

	repo := repository.NewPostgresUrlRepository(db)

	// Add cache dependency - direct Redis for simplicity
	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	// Use a minimal cache interface
	testCache := &testCache{client: rdb}

	svc := service.NewUrlService(repo, testCache, 12*time.Hour)

	now := time.Now().Add(1 * time.Hour)
	repo.Create(&domain.UrlMapping{
		ShortCode: "svc001",
		LongUrl:   "https://example.com/service",
		ExpiresAt: &now,
	})

	got, err := svc.GetByCode(context.Background(), "svc001")
	if err != nil {
		t.Fatalf("failed to get by code: %v", err)
	}
	if got.LongUrl != "https://example.com/service" {
		t.Errorf("expected https://example.com/service, got %s", got.LongUrl)
	}
}

type testCache struct {
	client *redis.Client
}

func (c *testCache) Get(ctx context.Context, code string) (*domain.UrlMapping, error) {
	return nil, nil
}

func (c *testCache) Set(ctx context.Context, url *domain.UrlMapping) error {
	return nil
}
