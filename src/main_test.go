package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServerPool_NextIndex ensures that the round-robin counter safely increments
// and wraps around the length of the slice properly.
func TestServerPool_NextIndex(t *testing.T) {
	u1, _ := url.Parse("http://localhost:8081")
	u2, _ := url.Parse("http://localhost:8082")

	pool := &ServerPool{
		backends: []*Backend{
			{URL: u1, Alive: true},
			{URL: u2, Alive: true},
		},
	}
	
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

	pool := &ServerPool{
		backends: []*Backend{b1, b2},
		current:  0,
	}

	peer := pool.GetNextPeer()
	if peer == nil {
		t.Fatal("Expected to find a peer, got nil")
	}

	if peer.URL.String() != "http://localhost:8082" {
		t.Errorf("Expected to skip dead backend and get u2, got %s", peer.URL)
	}
}

// TestLoadBalancer_MaxAttempts verifies that if a request context overflows
// the max attempts budget, it explicitly returns a 503 error.
func TestLoadBalancer_MaxAttempts(t *testing.T) {
	pool := &ServerPool{}
	lbInstance := &LoadBalancer{serverPool: pool}

	// Forge an incoming request already carrying an overloaded attempt context value
	req := httptest.NewRequest("GET", "http://localhost:3000/", nil)
	ctx := context.WithValue(req.Context(), Attempts, 4) // Max is 3
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	lbInstance.lb(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503, got %d", rec.Code)
	}
}