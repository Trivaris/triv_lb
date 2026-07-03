package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL				*url.URL
	Alive			bool
	mux				sync.RWMutex
	ReverseProxy	*httputil.ReverseProxy
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
  	b.mux.RUnlock()
  	return
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) isBackendAlive() bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", b.URL.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	_ = conn.Close()
	return true
}

func NewBackend(targetURL *url.URL, serverPool *ServerPool, lb http.HandlerFunc) *Backend {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	backend := &Backend{
		URL:          targetURL,
		Alive:        true,
		ReverseProxy: proxy,
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", targetURL.Host, e.Error())
		retries := GetRetryFromContext(request)
		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(request.Context(), Retry, retries+1)
			proxy.ServeHTTP(writer, request.WithContext(ctx))
			return
		}

		// After 3 retries, mark THIS specific backend as down
		backend.SetAlive(false) 

		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		
		// Fall back to the main load balancer loop to pick a different backend
		lb(writer, request.WithContext(ctx))
	}

	return backend
}
