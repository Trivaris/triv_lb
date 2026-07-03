package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	port := flag.String("port", "3000", "Port to run the load balancer on")
	backendList := flag.String("backends", "http://localhost:8081,http://localhost:8082,http://localhost:8083", "Comma-separated list of backend servers")
	flag.Parse()

	serverList := strings.Split(*backendList, ",")
	if len(serverList) == 0 || serverList[0] == "" {
		log.Fatal("You must provide at least one backend server")
	}

	pool := &ServerPool{}
	lbInstance := &LoadBalancer{serverPool: pool}

	var backends []*Backend
	for _, tok := range serverList {
		serverURL, err := url.Parse(strings.TrimSpace(tok))
		if err != nil {
			log.Fatalf("Invalid backend URL %s: %v", tok, err)
		}

		backend := NewBackend(serverURL, pool, lbInstance.lb)
		backends = append(backends, backend)
		log.Printf("Registered backend: %s\n", serverURL)
	}
	pool.backends = backends

	go pool.backgroundHealthCheck(20 * time.Second)

	server := http.Server{
		Addr:    ":" + *port,
		Handler: http.HandlerFunc(lbInstance.lb),
	}

	log.Printf("Load Balancer running on port %s...\n", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}