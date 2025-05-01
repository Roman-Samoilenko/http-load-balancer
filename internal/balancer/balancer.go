package balancer

import (
	"sync"
)

// Backend представляет бэкенд-сервер
type Backend struct {
	URL         string
	Weight      int
	ActiveConns int64
	IsAlive     bool
	mu          sync.RWMutex
}

// Balancer интерфейс для различных алгоритмов балансировки
type Balancer interface {
	NextBackend() *Backend
	AddBackend(backend *Backend)
	RemoveBackend(url string)
	MarkBackendDown(url string)
	MarkBackendUp(url string)
	Backends() []*Backend
}
