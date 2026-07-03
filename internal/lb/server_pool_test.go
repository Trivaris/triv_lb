package lb

import (
	"net/url"
	"testing"
)

// TestServerPool_NextIndex ensures that the round-robin counter safely increments
// and wraps around the length of the slice properly.
func TestServerPool_NextIndex(t *testing.T) {
	u1, _ := url.Parse("http://localhost:8081")
	u2, _ := url.Parse("http://localhost:8082")

	pool := NewServerPool()
	pool.SetBackends([]*Backend{
		{URL: u1, Alive: true},
		{URL: u2, Alive: true},
	})
	
	// Assert basic round-robin pacing: 1 -> 0 -> 1
	if idx := pool.NextIndex(); idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}
	if idx := pool.NextIndex(); idx != 0 {
		t.Errorf("Expected index 0, got %d", idx)
	}
	if idx := pool.NextIndex(); idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}
}

// TestServerPool_GetNextPeer_SkipsDead ensures that the pool loops through
// dead servers until it finds a live one.
func TestServerPool_GetNextPeer_SkipsDead(t *testing.T) {
	u1, _ := url.Parse("http://localhost:8081")
	u2, _ := url.Parse("http://localhost:8082")

	b1 := &Backend{URL: u1, Alive: false} // Dead
	b2 := &Backend{URL: u2, Alive: true}  // Alive


	pool := NewServerPool()
	pool.SetBackends([]*Backend{b1, b2})

	peer := pool.GetNextPeer()
	if peer == nil {
		t.Fatal("Expected to find a peer, got nil")
	}

	if peer.URL.String() != "http://localhost:8082" {
		t.Errorf("Expected to skip dead backend and get u2, got %s", peer.URL)
	}
}