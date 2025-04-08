package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/aleka07/go-digital-twin/pkg/messaging_sim"
	"github.com/aleka07/go-digital-twin/pkg/registry"
)

// Server represents the HTTP API server
type Server struct {
	Router   *chi.Mux
	Registry *registry.Registry
	PubSub   *messaging_sim.PubSub
	wg       sync.WaitGroup
}

// NewServer creates a new API server
func NewServer(reg *registry.Registry, pubsub *messaging_sim.PubSub) *Server {
	s := &Server{
		Router:   chi.NewRouter(),
		Registry: reg,
		PubSub:   pubsub,
	}

	// Set up middleware
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.Timeout(30 * time.Second))

	// Register routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	// Twin management
	s.Router.Route("/twins", func(r chi.Router) {
		r.Post("/", s.CreateTwin)
		r.Get("/", s.ListTwins)
		
		r.Route("/{twinID}", func(r chi.Router) {
			r.Get("/", s.GetTwin)
			r.Put("/", s.UpdateTwin)
			r.Delete("/", s.DeleteTwin)
			
			// Feature management
			r.Route("/features", func(r chi.Router) {
				r.Get("/", s.GetFeatures)
				
				r.Route("/{featureID}", func(r chi.Router) {
					r.Get("/", s.GetFeature)
					r.Put("/", s.UpdateFeature)
					r.Delete("/", s.DeleteFeature)
					
					// Property management
					r.Route("/properties", func(r chi.Router) {
						r.Get("/", s.GetProperties)
						r.Put("/", s.UpdateProperties)
						
						r.Route("/{propKey}", func(r chi.Router) {
							r.Get("/", s.GetProperty)
							r.Put("/", s.UpdateProperty)
							r.Delete("/", s.DeleteProperty)
						})
					})
				})
			})
		})
	})

	// Health check
	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	return server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Wait for all in-flight requests to complete
	waitCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
