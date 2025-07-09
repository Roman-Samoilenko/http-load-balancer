package balancer

import (
	"sync"
	"sync/atomic"

	"github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
)

// RoundRobin реализует алгоритм балансировки Round-Robin
type RoundRobin struct {
	backends []*Backend
	current  uint32 // Индекс текущего бэкенда
	mu       sync.RWMutex
}

// NewRoundRobin создает новый экземпляр балансировщика Round-Robin
func NewRoundRobin(backends []*Backend) *RoundRobin {
	return &RoundRobin{
		backends: backends,
		current:  0,
	}
}

// NextBackend возвращает следующий бэкенд для обработки запроса
// Реализует основную логику алгоритма Round-Robin
func (r *RoundRobin) NextBackend() *Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Если нет доступных бэкендов, возвращаем nil
	if len(r.backends) == 0 {
		return nil
	}

	// Проходим по всем бэкендам, начиная со следующего после текущего
	next := atomic.AddUint32(&r.current, 1) % uint32(len(r.backends))

	// Проверяем все бэкенды, начиная с next
	for i := uint32(0); i < uint32(len(r.backends)); i++ {
		idx := (next + i) % uint32(len(r.backends))
		backend := r.backends[idx]

		// Проверяем, что бэкенд доступен
		backend.mu.RLock()
		isAlive := backend.IsAlive
		backend.mu.RUnlock()

		if isAlive {
			// Инкрементируем счетчик соединений для выбранного бэкенда
			backend.IncrementConnections()
			return backend
		}
	}

	// Если нет доступных бэкендов, возвращаем nil
	return nil
}

// AddBackend добавляет новый бэкенд в пул
func (r *RoundRobin) AddBackend(backend *Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Устанавливаем флаг активности по умолчанию
	backend.SetAlive(true)
	r.backends = append(r.backends, backend)
}

// RemoveBackend удаляет бэкенд из пула по URL
func (r *RoundRobin) RemoveBackend(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, backend := range r.backends {
		if backend.URL == url {
			// Удаляем бэкенд из слайса
			r.backends = append(r.backends[:i], r.backends[i+1:]...)
			return
		}
	}
}

// MarkBackendDown помечает бэкенд как недоступный
func (r *RoundRobin) MarkBackendDown(url string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, backend := range r.backends {
		if backend.URL == url {
			backend.SetAlive(false)
			return
		}
	}
}

// MarkBackendUp помечает бэкенд как доступный
func (r *RoundRobin) MarkBackendUp(url string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, backend := range r.backends {
		if backend.URL == url {
			backend.SetAlive(true)
			return
		}
	}
}

// Backends возвращает список всех бэкендов
func (r *RoundRobin) Backends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.backends
}
