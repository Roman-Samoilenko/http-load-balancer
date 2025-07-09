package balancer

import (
	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
	"sync"
)

// Balancer интерфейс для различных алгоритмов балансировки
type Balancer interface {
	NextBackend() *Backend
	AddBackend(backend *Backend)
	RemoveBackend(url string)
	MarkBackendDown(url string)
	MarkBackendUp(url string)
	Backends() []*Backend
}

type BaseBalancer struct {
	backends []*Backend
	mu       sync.RWMutex
}

// AddBackend добавляет новый бэкенд в пул
func (bb *BaseBalancer) AddBackend(backend *Backend) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	// Устанавливаем флаг активности по умолчанию
	backend.SetAlive(true)
	bb.backends = append(bb.backends, backend)
}

// RemoveBackend удаляет бэкенд из пула по URL
func (bb *BaseBalancer) RemoveBackend(url string) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	for i, backend := range bb.backends {
		if backend.URL == url {
			// Удаляем бэкенд из слайса
			bb.backends = append(bb.backends[:i], bb.backends[i+1:]...)
			return
		}
	}
}

// MarkBackendDown помечает бэкенд как недоступный
func (bb *BaseBalancer) MarkBackendDown(url string) {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	for _, backend := range bb.backends {
		if backend.URL == url {
			backend.SetAlive(false)
			return
		}
	}
}

// MarkBackendUp помечает бэкенд как доступный
func (bb *BaseBalancer) MarkBackendUp(url string) {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	for _, backend := range bb.backends {
		if backend.URL == url {
			backend.SetAlive(true)
			return
		}
	}
}

// Backends возвращает список всех бэкендов
func (bb *BaseBalancer) Backends() []*Backend {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	return bb.backends
}
