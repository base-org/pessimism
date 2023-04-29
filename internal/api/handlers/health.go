package handlers

import (
	"net/http"

	"github.com/go-chi/render"
)

// HealthCheck ... Handle health check
func (ph *PessimismHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, ph.service.CheckHealth())
}
