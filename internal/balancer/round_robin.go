package balancer

import (
	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
	"sync/atomic"
)

// RoundRobin реализует алгоритм балансировки Round-Robin
type RoundRobin struct {
	BaseBalancer
	current uint32 // Индекс текущего бэкенда
}

// NewRoundRobin создает новый экземпляр балансировщика Round-Robin
func NewRoundRobin(backends []*Backend) *RoundRobin {
	return &RoundRobin{
		BaseBalancer: BaseBalancer{
			backends: backends,
		},
		current: 0,
	}
}

// NextBackend возвращает следующий бэкенд для обработки запроса
// реализует основную логику алгоритма Round-Robin
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
		backend.Mu.RLock()
		isAlive := backend.IsAlive
		backend.Mu.RUnlock()

		if isAlive {
			// Инкрементируем счетчик соединений для выбранного бэкенда
			backend.IncrementConnections()
			return backend
		}
	}

	// Если нет доступных бэкендов, возвращаем nil
	return nil
}
