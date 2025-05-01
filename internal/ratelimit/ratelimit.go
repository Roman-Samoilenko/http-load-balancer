package ratelimit

import (
	"sync"
	"time"
)

// NewTokenBucket создает новый экземпляр TokenBucket
func NewTokenBucket(capacity int, rate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Allow проверяет и берет токен, если доступен
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.refill(now)

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// refill пополняет токены в бакете
func (tb *TokenBucket) refill(now time.Time) {
	delta := now.Sub(tb.lastRefill).Seconds() // время, прошедшее с последнего пополнения
	tb.lastRefill = now                       // обновляет время последнего пополнения

	// Вычисляем, сколько токенов нужно добавить
	newTokens := delta * tb.rate

	// Обновляем количество токенов, не превышая емкость
	if newTokens > 0 {
		tb.tokens = min(float64(tb.capacity), tb.tokens+newTokens)
	}
}

// Manager управляет клиентами и их rate limits
type Manager struct {
	clients     map[string]*Client
	defaultCap  int
	defaultRate float64
	mu          sync.RWMutex
}

// NewManager создает новый экземпляр Manager
func NewManager(defaultCap int, defaultRate float64) *Manager {
	return &Manager{
		clients:     make(map[string]*Client),
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
	}
}

// GetClient возвращает клиента по ID (IP или API-ключ)
func (m *Manager) GetClient(id string) *Client {
	m.mu.RLock()
	client, exists := m.clients[id]
	m.mu.RUnlock()

	if exists {
		return client
	}

	// Если клиент не существует, создаем нового
	m.mu.Lock()
	defer m.mu.Unlock()

	// Повторная проверка после получения эксклюзивной блокировки
	if client, exists = m.clients[id]; exists {
		return client
	}

	// Создание нового клиента
	client = &Client{
		ID:     id,
		Bucket: NewTokenBucket(m.defaultCap, m.defaultRate),
	}
	m.clients[id] = client

	return client
}

// Allow проверяет разрешение запроса для клиента
func (m *Manager) Allow(id string) bool {
	client := m.GetClient(id)
	return client.Bucket.Allow()
}
