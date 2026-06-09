package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suman2280/go-url-shortener/internal/service"
	"github.com/suman2280/go-url-shortener/pkg/analytics"
)

type Handler struct {
	urlService *service.UrlService
	analytics  *analytics.Worker
	baseURL    string
}

func NewHandler(us *service.UrlService, aw *analytics.Worker, baseURL string) *Handler {
	return &Handler{urlService: us, analytics: aw, baseURL: baseURL}
}

type ShortenRequest struct {
	LongUrl        string `json:"long_url" binding:"required"`
	Alias          string `json:"alias,omitempty"`
	ExpiresInHours *int   `json:"expires_in_hours,omitempty"`
}

type ShortenResponse struct {
	ShortCode string  `json:"short_code"`
	ShortURL  string  `json:"short_url"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatsResponse struct {
	ShortCode     string `json:"short_code"`
	LongUrl       string `json:"long_url"`
	ClickCount    int64  `json:"click_count"`
	ExpiresAt     string `json:"expires_at"`
	IsCustomAlias bool   `json:"is_custom_alias"`
	CreatedAt     string `json:"created_at"`
}

// CreateShortUrl handles POST /api/shorten
// @Summary Create a shortened URL
// @Tags urls
// @Accept json
// @Produce json
// @Param request body ShortenRequest true "URL to shorten"
// @Success 201 {object} ShortenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/shorten [post]
func (h *Handler) CreateShortUrl(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var alias *string
	if req.Alias != "" {
		alias = &req.Alias
	}

	url, err := h.urlService.CreateShortUrl(req.LongUrl, alias, req.ExpiresInHours)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "alias already taken" {
			status = http.StatusConflict
		}
		c.JSON(status, ErrorResponse{Error: err.Error()})
		return
	}

	var expiresAt *string
	if url.ExpiresAt != nil {
		e := url.ExpiresAt.Format(time.RFC3339)
		expiresAt = &e
	}

	c.JSON(http.StatusCreated, ShortenResponse{
		ShortCode: url.ShortCode,
		ShortURL:  h.baseURL + "/" + url.ShortCode,
		ExpiresAt: expiresAt,
	})
}

// HandleShortUrlRedirect handles GET /:code
// @Summary Redirect to the original URL
// @Tags urls
// @Produce json
// @Param code path string true "Short code"
// @Success 301 {string} string "Redirect"
// @Failure 404 {object} ErrorResponse
// @Failure 410 {object} ErrorResponse
// @Router /{code} [get]
func (h *Handler) HandleShortUrlRedirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	if h.urlService.IsExpired(url) {
		c.JSON(http.StatusGone, ErrorResponse{Error: "URL has expired"})
		return
	}

	go func() {
		_ = h.analytics.PublishClick(code)
		_ = h.urlService.RecordClick(code)
	}()

	c.Redirect(http.StatusMovedPermanently, url.LongUrl)
}

// GetUrlMeta handles GET /api/:code
// @Summary Get URL metadata
// @Tags urls
// @Produce json
// @Param code path string true "Short code"
// @Success 200 {object} ShortenResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/{code} [get]
func (h *Handler) GetUrlMeta(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	expired := h.urlService.IsExpired(url)

	var expiresAt *string
	if url.ExpiresAt != nil {
		e := url.ExpiresAt.Format(time.RFC3339)
		expiresAt = &e
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code": url.ShortCode,
		"short_url":  h.baseURL + "/" + url.ShortCode,
		"expires_at": expiresAt,
		"expired":    expired,
	})
}

// GetStats handles GET /api/:code/stats
// @Summary Get click statistics for a short URL
// @Tags urls
// @Produce json
// @Param code path string true "Short code"
// @Success 200 {object} StatsResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/{code}/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetStats(code)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StatsResponse{
		ShortCode:     url.ShortCode,
		LongUrl:       url.LongUrl,
		ClickCount:    url.ClickCount,
		ExpiresAt:     url.ExpiresAt.Format(time.RFC3339),
		IsCustomAlias: url.IsCustomAlias,
		CreatedAt:     url.CreatedAt.Format(time.RFC3339),
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
