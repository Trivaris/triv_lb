package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/trivaris/triv_lb/internal/lb"
	"github.com/trivaris/triv_lb/internal/config"
)

func main() {
	cfgPath := flag.String("config", config.GetSystemConfigPath(), "Path to the configuration file")
	flagPort := flag.String("port", "", "Port to run the load balancer on")
	flag.Parse()

	if err := config.EnsureConfigFile(*cfgPath); err != nil {
		log.Fatalf("Critical error initializing config file: %v", err)
	}

	cfg := &config.Config{}
	if err := cfg.LoadFromFile(*cfgPath); err != nil {
		log.Fatalf("Critical error loading config file: %v", err)
	}

	if *flagPort != "" {
		cfg.Port = *flagPort
	}

	pool := &lb.ServerPool{}
	lbInstance := lb.NewLoadBalancer(pool)

	var backends []*lb.Backend
	for _, tok := range cfg.BackendList {
		serverURL, err := url.Parse(strings.TrimSpace(tok))
		if err != nil {
			log.Fatalf("Invalid backend URL %s: %v", tok, err)
		}

		backend := lb.NewBackend(serverURL, pool, lbInstance.LoadBalance)
		backends = append(backends, backend)
		log.Printf("Registered backend: %s\n", serverURL)
	}

	pool.SetBackends(backends)

	go pool.BackgroundHealthCheck(20 * time.Second)

	server := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: http.HandlerFunc(lbInstance.LoadBalance),
	}

	log.Printf("Load Balancer running on port %s...\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}
