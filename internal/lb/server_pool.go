package lb

import (
	"log"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	backends []*Backend
	current uint64
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: []*Backend{},
		current:  0,
	}
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) SetBackends(backends []*Backend) {
	s.backends = backends
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := next + len(s.backends) // full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
      	// next has already been stored in current, due to s.NextIndex(), so don't store it again
	  	if i != next {
        	atomic.StoreUint64(&s.current, uint64(idx))
      	}
      	return s.backends[idx]
    	}
  	}
  	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := b.isBackendAlive()
		b.SetAlive(alive)
		if !alive {
		status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func (s *ServerPool) BackgroundHealthCheck(d time.Duration) {
	t := time.NewTicker(d)
	for range t.C {
		log.Println("Starting health check...")
		s.HealthCheck()
		log.Println("Health check completed")
	}
}