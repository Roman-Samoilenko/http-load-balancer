package balancer

import (
	"sync"

	"github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
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
