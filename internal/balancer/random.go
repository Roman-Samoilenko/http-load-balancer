package balancer

import (
	. "github.com/Roman-Samoilenko/http-load-balancer/internal/backend"
	"math/rand/v2"
)

type Random struct {
	BaseBalancer
}

func NewRandom(backends []*Backend) *Random {
	return &Random{
		BaseBalancer: BaseBalancer{
			backends: backends,
		},
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
