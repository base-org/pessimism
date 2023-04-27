package handlers

import (
	"net/http"

	"github.com/base-org/pessimism/internal/api/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handlers interface {
	HealthCheck(w http.ResponseWriter, r *http.Request)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// PessimismHandler ... Server handler logic
type PessimismHandler struct {
	service service.Service
	router  *chi.Mux
}

// New ... Initializer
func New(service service.Service) (Handlers, error) {
	handlers := &PessimismHandler{service: service}
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	registerEndpoint("/health", router.Get, handlers.HealthCheck)
	handlers.router = router

	return handlers, nil
}

// ServeHTTP serves a http request given a response builder and request
func (ph *PessimismHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.router.ServeHTTP(w, r)
}

// registerEndpoint registers an endpoint to the router for a specified method type and handlerFunction
func registerEndpoint(endpoint string, routeMethod func(pattern string, handlerFn http.HandlerFunc), handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	routeMethod(endpoint, http.HandlerFunc(handlerFunc).ServeHTTP)
}
