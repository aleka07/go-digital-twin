package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleka07/go-digital-twin/pkg/api"
	"github.com/aleka07/go-digital-twin/pkg/messaging_sim"
	"github.com/aleka07/go-digital-twin/pkg/registry"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Create components
	reg := registry.NewRegistry()
	pubsub := messaging_sim.NewPubSub()
	server := api.NewServer(reg, pubsub)

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	serverAddr := fmt.Sprintf("0.0.0.0:%d", *port)
	go func() {
		log.Printf("Starting Digital Twin server on %s", serverAddr)
		if err := server.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Subscribe to events for logging
	eventCh := pubsub.Subscribe("twin.+")
	go func() {
		for event := range eventCh {
			log.Printf("Event: %s - %v", event.Topic, event.Payload)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	// Close pubsub
	pubsub.Close()

	log.Println("Server gracefully stopped")
}
