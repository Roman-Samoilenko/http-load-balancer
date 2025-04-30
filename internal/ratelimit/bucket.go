package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket реализует алгоритм Token Bucket
type TokenBucket struct {
	capacity   int     // Максимальное количество токенов
	tokens     float64 // Текущее количество токенов
	rate       float64 // Скорость пополнения токенов в секунду
	lastRefill time.Time
	mu         sync.Mutex
}

//// NewTokenBucket создает новый экземпляр TokenBucket
//func NewTokenBucket(capacity int, rate float64) *TokenBucket {
//	// Создание Token Bucket
//}
//
//// Allow проверяет и берет токен, если доступен
//func (tb *TokenBucket) Allow() bool {
//	// Проверка наличия токена и его взятие
//}
