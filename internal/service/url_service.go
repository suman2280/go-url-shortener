package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/suman2280/go-url-shortener/internal/domain"
	"github.com/suman2280/go-url-shortener/internal/repository"
	"github.com/suman2280/go-url-shortener/internal/shortener"
	"github.com/suman2280/go-url-shortener/pkg/validator"
)

type UrlService struct {
	repo          repository.UrlRepository
	cache         cacheStore
	defaultExpiry time.Duration
}

type cacheStore interface {
	Get(ctx context.Context, code string) (*domain.UrlMapping, error)
	Set(ctx context.Context, url *domain.UrlMapping) error
}

func NewUrlService(repo repository.UrlRepository, cache cacheStore, defaultExpiry time.Duration) *UrlService {
	return &UrlService{repo: repo, cache: cache, defaultExpiry: defaultExpiry}
}

func (s *UrlService) CreateShortUrl(longUrl string, alias *string, expiresInHours *int) (*domain.UrlMapping, error) {
	if !validator.IsValidURL(longUrl) {
		return nil, fmt.Errorf("invalid URL format")
	}
	if len(longUrl) > 2048 {
		return nil, fmt.Errorf("URL exceeds maximum length of 2048 characters")
	}

	var code string
	isCustom := false

	if alias != nil && *alias != "" {
		if !validator.IsValidAlias(*alias) {
			return nil, fmt.Errorf("invalid alias: must be 1-50 alphanumeric characters or hyphens")
		}
		exists, err := s.repo.CodeExists(*alias)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("alias already taken")
		}
		code = *alias
		isCustom = true
	} else {
		var err error
		code, err = s.generateUniqueCode()
		if err != nil {
			return nil, err
		}
	}

	var expiry *time.Time
	if expiresInHours != nil && *expiresInHours > 0 {
		t := time.Now().Add(time.Duration(*expiresInHours) * time.Hour)
		expiry = &t
	} else {
		t := time.Now().Add(s.defaultExpiry)
		expiry = &t
	}

	mapping := &domain.UrlMapping{
		ShortCode:     code,
		LongUrl:       longUrl,
		ExpiresAt:     expiry,
		IsCustomAlias: isCustom,
	}

	if err := s.repo.Create(mapping); err != nil {
		return nil, fmt.Errorf("failed to save URL mapping: %w", err)
	}

	return mapping, nil
}

func (s *UrlService) GetByCode(ctx context.Context, code string) (*domain.UrlMapping, error) {
	cached, err := s.cache.Get(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("cache error: %w", err)
	}
	if cached != nil {
		return cached, nil
	}

	mapping, err := s.repo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("short URL not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := s.cache.Set(ctx, mapping); err != nil {
		return nil, fmt.Errorf("cache set error: %w", err)
	}

	return mapping, nil
}

func (s *UrlService) IsExpired(mapping *domain.UrlMapping) bool {
	return mapping.ExpiresAt != nil && time.Now().After(*mapping.ExpiresAt)
}

func (s *UrlService) GetStats(code string) (*domain.UrlMapping, error) {
	mapping, err := s.repo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("short URL not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return mapping, nil
}

func (s *UrlService) RecordClick(code string) error {
	return s.repo.IncrementClickCount(code)
}

func (s *UrlService) CleanupExpired() error {
	return s.repo.DeleteExpired()
}

func (s *UrlService) generateUniqueCode() (string, error) {
	for i := 0; i < shortener.MaxRetries; i++ {
		code, err := shortener.GenerateCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate code: %w", err)
		}
		exists, err := s.repo.CodeExists(code)
		if err != nil {
			return "", fmt.Errorf("failed to check code: %w", err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", shortener.ErrMaxRetries{}
}
