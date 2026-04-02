package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/snowmerak/arena/gateway"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	cfg, err := gateway.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gw, err := gateway.NewGateway(cfg)
	if err != nil {
		log.Fatalf("failed to initialize gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gw) 

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s with %d models configured", addr, len(cfg.Models))
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
