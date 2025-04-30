package ratelimit

import (
	"time"
)

// Client представляет клиента с настройками rate limiting
type Client struct {
	ID        string
	Bucket    *TokenBucket
	Capacity  int
	Rate      float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

//// NewClient создает нового клиента
//func NewClient(id string, capacity int, rate float64) *Client {
//	// Создание клиента
//}
