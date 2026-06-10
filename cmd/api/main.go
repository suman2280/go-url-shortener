package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/suman2280/go-url-shortener/docs"
	"github.com/suman2280/go-url-shortener/internal/cache"
	"github.com/suman2280/go-url-shortener/internal/config"
	"github.com/suman2280/go-url-shortener/internal/domain"
	"github.com/suman2280/go-url-shortener/internal/handler"
	"github.com/suman2280/go-url-shortener/internal/middleware"
	"github.com/suman2280/go-url-shortener/internal/repository"
	"github.com/suman2280/go-url-shortener/internal/service"
	"github.com/suman2280/go-url-shortener/pkg/analytics"
)

// @title URL Shortener API
// @version 1.0
// @description A production-ready URL shortener with caching, analytics, and observability.
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&domain.UrlMapping{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.CacheTTL)
	ctx := context.Background()
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	repo := repository.NewPostgresUrlRepository(db)
	urlService := service.NewUrlService(repo, redisCache, cfg.DefaultExpiry)

	analyticsWorker := analytics.NewWorker(
		redisCache.Client(),
		func(code string) error {
			return urlService.RecordClick(code)
		},
		func() error {
			return urlService.CleanupExpired()
		},
	)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	go analyticsWorker.Run(workerCtx)

	h := handler.NewHandler(urlService, analyticsWorker, "http://localhost:8080")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.PrometheusMiddleware())

	r.StaticFile("/", "./static/index.html")

	r.GET("/health", h.HealthCheck)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/api/shorten", middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst), h.CreateShortUrl)
	r.GET("/api/:code", h.GetUrlMeta)
	r.GET("/api/:code/stats", h.GetStats)

	r.GET("/:code", h.HandleShortUrlRedirect)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := r.Run(":" + cfg.ServerPort); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")
	workerCancel()
	if err := redisCache.Close(); err != nil {
		log.Printf("error closing redis: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("error getting sql db: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}

	log.Println("server exited")
}
