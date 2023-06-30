package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Config ... Server configuration options
type Config struct {
	Host            string
	Port            int
	KeepAlive       int
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

// Server ... Server representation struct
type Server struct {
	Cfg        *Config
	serverHTTP *http.Server
}

// New ... Initializer
func New(ctx context.Context, cfg *Config, apiHandlers handlers.Handlers) (*Server, func(), error) {
	restServer := initializeServer(cfg, apiHandlers)

	stop := func() {
		logging.WithContext(ctx).Info("starting to shutdown REST API HTTP server")

		ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimeout)*time.Second)
		if err := restServer.serverHTTP.Shutdown(ctx); err != nil {
			logging.WithContext(ctx).Error("failed to shutdown REST API HTTP server")
			panic(err)
		}

		defer cancel()
	}

	return restServer, stop, nil
}

// spawnServer ... Starts a listen and serve API routine
func spawnServer(server *Server) {
	logging.NoContext().Info("Starting REST API HTTP server",
		zap.String("address", server.serverHTTP.Addr))

	if err := server.serverHTTP.ListenAndServe(); err != http.ErrServerClosed {
		logging.NoContext().Error("failed to run REST API HTTP server", zap.String("address", server.serverHTTP.Addr))
		panic(err)
	}
}

func (s *Server) Start() {
	go spawnServer(s)
}

// initializeServer ... Initializes server struct object
func initializeServer(config *Config, handler http.Handler) *Server {
	return &Server{
		Cfg: config,
		serverHTTP: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:           handler,
			IdleTimeout:       time.Duration(config.KeepAlive) * time.Second,
			ReadHeaderTimeout: time.Duration(config.ReadTimeout) * time.Second,
			ReadTimeout:       time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:      time.Duration(config.WriteTimeout) * time.Second,
		},
	}
}
