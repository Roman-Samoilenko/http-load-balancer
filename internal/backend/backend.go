package backend

import (
	"sync"
	"sync/atomic"
)

// Backend представляет бэкенд-сервер
type Backend struct {
	URL         string
	Weight      int
	ActiveConns int64
	IsAlive     bool
	Mu          sync.RWMutex
}

// IncrementConnections увеличивает счетчик активных соединений
func (b *Backend) IncrementConnections() {
	atomic.AddInt64(&b.ActiveConns, 1)
}

// DecrementConnections уменьшает счетчик активных соединений
func (b *Backend) DecrementConnections() {
	atomic.AddInt64(&b.ActiveConns, -1)
}

// SetAlive устанавливает статус доступности бэкенда
func (b *Backend) SetAlive(isAlive bool) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.IsAlive = isAlive
}

// GetActiveConnections возвращает текущее количество активных соединений
func (b *Backend) GetActiveConnections() int64 {
	return atomic.LoadInt64(&b.ActiveConns)
}
