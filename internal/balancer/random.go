package balancer

import (
	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
	"math/rand/v2"
	"sync"
)

type Random struct {
	backends []*Backend
	mu       sync.RWMutex
}

func NewRandom(backends []*Backend) *Random {
	return &Random{
		backends: backends,
	}
}

func (r *Random) NextBackend() *Backend {
	r.mu.Lock()
	defer r.mu.Unlock()

	backAlive := make([]*Backend, 0)

	for _, b := range r.backends {
		b.Mu.RLock()
		if b.IsAlive {
			backAlive = append(backAlive, b)
		}
		b.Mu.RUnlock()
	}

	if len(backAlive) == 0 {
		return nil
	}

	indexBE := rand.N(len(backAlive))
	backAlive[indexBE].IncrementConnections()
	return backAlive[indexBE]

}

func (r *Random) AddBackend(backend *Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()

	backend.SetAlive(true)
	r.backends = append(r.backends, backend)
	return
}

func (r *Random) RemoveBackend(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, b := range r.backends {
		if b.URL == url {
			r.backends = append(r.backends[:i], r.backends[i+1:]...)
			return
		}
	}
}

func (r *Random) MarkBackendDown(url string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.backends {
		if b.URL == url {
			b.SetAlive(false)
			return
		}
	}
}

func (r *Random) MarkBackendUp(url string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.backends {
		if b.URL == url {
			b.SetAlive(true)
			return
		}
	}
}

func (r *Random) Backends() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.backends
}
