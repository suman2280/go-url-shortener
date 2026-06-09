package repository

import (
	"github.com/suman2280/go-url-shortener/internal/domain"
)

type UrlRepository interface {
	Create(url *domain.UrlMapping) error
	GetByCode(code string) (*domain.UrlMapping, error)
	IncrementClickCount(code string) error
	DeleteExpired() error
	CodeExists(code string) (bool, error)
}
