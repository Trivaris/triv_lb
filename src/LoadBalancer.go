package main

import (
	"log"
	"net/http"
)

type LoadBalancer struct {
	serverPool	*ServerPool
}

func (lb *LoadBalancer) lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)

	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	
	peer := lb.serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}