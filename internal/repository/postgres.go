package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/suman2280/go-url-shortener/internal/domain"
)

type PostgresUrlRepository struct {
	db *gorm.DB
}

func NewPostgresUrlRepository(db *gorm.DB) *PostgresUrlRepository {
	return &PostgresUrlRepository{db: db}
}

func (r *PostgresUrlRepository) Create(url *domain.UrlMapping) error {
	return r.db.Create(url).Error
}

func (r *PostgresUrlRepository) GetByCode(code string) (*domain.UrlMapping, error) {
	var url domain.UrlMapping
	err := r.db.Where("short_code = ?", code).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *PostgresUrlRepository) IncrementClickCount(code string) error {
	return r.db.Model(&domain.UrlMapping{}).
		Where("short_code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).
		Error
}

func (r *PostgresUrlRepository) DeleteExpired() error {
	return r.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&domain.UrlMapping{}).
		Error
}

func (r *PostgresUrlRepository) CodeExists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.UrlMapping{}).
		Where("short_code = ?", code).
		Count(&count).
		Error
	return count > 0, err
}
