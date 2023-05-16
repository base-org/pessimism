package handlers

import (
	"context"
	"net/http"

	pess_middleware "github.com/base-org/pessimism/internal/api/handlers/middleware"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/go-chi/chi"
	chi_middleware "github.com/go-chi/chi/middleware"
)

type Handlers interface {
	HealthCheck(w http.ResponseWriter, r *http.Request)
	RunInvariant(w http.ResponseWriter, r *http.Request)

	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// PessimismHandler ... Server handler logic
type PessimismHandler struct {
	ctx     context.Context
	service service.Service
	router  *chi.Mux
}

type Route = string

const (
	healthRoute    = "/health"
	invariantRoute = "/v0/invariant"
)

// New ... Initializer
func New(ctx context.Context, service service.Service) (Handlers, error) {
	handlers := &PessimismHandler{ctx: ctx, service: service}
	router := chi.NewRouter()

	router.Use(chi_middleware.Recoverer)

	router.Use(pess_middleware.InjectedLogging(logging.NoContext()))

	registerEndpoint(healthRoute, router.Get, handlers.HealthCheck)
	registerEndpoint(invariantRoute, router.Post, handlers.RunInvariant)

	handlers.router = router

	return handlers, nil
}

// ServeHTTP ... Serves a http request given a response builder and request
func (ph *PessimismHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.router.ServeHTTP(w, r)
}

// registerEndpoint ... Registers an endpoint to the router for a specified method type and handlerFunction
func registerEndpoint(endpoint string, routeMethod func(pattern string, handlerFn http.HandlerFunc),
	handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	routeMethod(endpoint, http.HandlerFunc(handlerFunc).ServeHTTP)
}
