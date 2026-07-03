package lb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoadBalancer_MaxAttempts verifies that if a request context overflows
// the max attempts budget, it explicitly returns a 503 error.
func TestLoadBalancer_MaxAttempts(t *testing.T) {
	pool := NewServerPool()
	lbInstance := NewLoadBalancer(pool)

	// Forge an incoming request already carrying an overloaded attempt context value
	req := httptest.NewRequest("GET", "http://localhost:3000/", nil)
	ctx := context.WithValue(req.Context(), Attempts, 4) // Max is 3
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	lbInstance.LoadBalance(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503, got %d", rec.Code)
	}
}