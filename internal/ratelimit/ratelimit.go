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
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Вычисляем, сколько токенов нужно добавить
	newTokens := elapsed * tb.rate

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

// AddClient добавляет нового клиента с указанными параметрами
func (m *Manager) AddClient(id string, capacity int, rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[id] = &Client{
		ID:     id,
		Bucket: NewTokenBucket(capacity, rate),
	}
}

// UpdateClient обновляет настройки клиента
func (m *Manager) UpdateClient(id string, capacity int, rate float64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[id]; !exists {
		return false
	}

	m.clients[id].Bucket = NewTokenBucket(capacity, rate)
	return true
}

// DeleteClient удаляет клиента
func (m *Manager) DeleteClient(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[id]; !exists {
		return false
	}

	delete(m.clients, id)
	return true
}

// Вспомогательная функция для Go < 1.21
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
