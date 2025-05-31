// Package server provides HTTP server functionality for the application.
// It implements the server interface and provides methods for server lifecycle management.
package server

import (
	"chief-checker/internal/service/server/serverInterface"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// server implements the serverInterface.Server interface.
// It provides HTTP server functionality with graceful shutdown support.
type server struct {
	handler serverInterface.Handler
	server  *http.Server
	done    chan struct{}
	mu      sync.Mutex
}

// NewServerHandler creates a new server instance with the provided handler.
// It returns an implementation of the serverInterface.Server interface.
func NewServerHandler(handler serverInterface.Handler) serverInterface.Server {
	return &server{
		handler: handler,
		mu:      sync.Mutex{},
		done:    make(chan struct{}),
	}
}

// StartServer initializes and starts the HTTP server on the specified port.
// It sets up routes and begins listening for requests.
// The server runs until it receives a stop signal.
func (s *server) StartServer(port string) {
	mux := http.NewServeMux()
	s.registerHandlers(mux)

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("Сервер запущен: http://localhost:%s", port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	<-s.done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}

	close(s.done)
}

// StopServer handles the server shutdown request.
// It returns a status response and initiates graceful shutdown.
func (s *server) StopServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})

	go func() {
		if err := s.processStop(context.Background()); err != nil {
			log.Printf("Ошибка при остановке сервера: %v", err)
		}
	}()
}

// processStop handles the server shutdown process.
// It ensures thread-safe shutdown of the server.
func (s *server) processStop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.done <- struct{}{}
		return nil
	}
	return nil
}

// Done returns a channel that is closed when the server is stopped.
// It can be used to wait for server shutdown.
func (s *server) Done() <-chan struct{} {
	return s.done
}

// registerHandlers sets up the HTTP routes for the server.
// It registers handlers for balance, addresses, and server control endpoints.
func (s *server) registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/balance", s.handler.BalanceData)
	mux.HandleFunc("/api/addresses", s.handler.Addresses)
	mux.HandleFunc("/api/stop", s.StopServer)
	mux.Handle("/", http.FileServer(http.Dir("web")))
}
