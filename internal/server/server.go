package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
)

// APIServer обёртка над http.Server для управления жизненным циклом сервера
type APIServer struct {
	httpServer *http.Server
}

func NewAPIServer(handler http.Handler) *APIServer {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	return &APIServer{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}
}

func (s *APIServer) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *APIServer) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	switch err := ctx.Err(); {
	case errors.Is(err, context.DeadlineExceeded):
		// timeout shutting down server
		return err
	case err == nil:
		// shutdown completed before timeout
		return nil
	default:
		// shutdown ended with error
		return err
	}
}
