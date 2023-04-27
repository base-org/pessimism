package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Config ... Server configuration options
type Config struct {
	Host            string
	Port            int
	ListenLimit     int
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
	go spawnServer(restServer)

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

// spawnServer ... Starts a counterparty listen and serve API routine
func spawnServer(server *Server) {
	logging.NoContext().Info("Starting REST API HTTP server",
		zap.String("adress", server.serverHTTP.Addr))

	if err := server.serverHTTP.ListenAndServe(); err != http.ErrServerClosed {

		logging.NoContext().Error("failed to run REST API HTTP server", zap.String("address", server.serverHTTP.Addr))
		panic(err)
	}
}

// initializeServer ... Initializes server struct object
func initializeServer(config *Config, handler http.Handler) *Server {

	return &Server{
		Cfg: config,
		serverHTTP: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(10) * time.Second,
			WriteTimeout: time.Duration(10) * time.Second,
		},
	}
}

// returns a channel to handle shutdown
func (sv *Server) done() <-chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

// waits for shutdown signal from shutdown cahnnel
func (sv *Server) Stop(stop func()) {
	done := <-sv.done()
	logging.NoContext().Info("Received shutdown OS signal", zap.String("signal", done.String()))
	stop()
	os.Exit(0)
}
