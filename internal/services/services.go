package services

import "sliding-window-rate-limiter/internal/models"

//go:generate mockgen -source=services.go -destination=./services_mock/services_mock.go -package=services_mock

// CounterServiceInterface handles the counter
type CounterServiceInterface interface {
	Hit() int64
	Count() int64
	Window() []models.Entry
}

// RateLimiterInterface handles the rate limiting part and global counter
type RateLimiterInterface interface {
	Hit(ipAddr string) (int64, int64, bool)
	Dump() error
}
