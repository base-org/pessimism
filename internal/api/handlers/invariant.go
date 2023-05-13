package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

func renderInvariantResponse(w http.ResponseWriter, r *http.Request,
	ir *models.InvResponse) {
	w.WriteHeader(ir.Code)
	render.JSON(w, r, ir)
}

// RunInvariant ... Handle invariant run request
func (ph *PessimismHandler) RunInvariant(w http.ResponseWriter, r *http.Request) {
	var body models.InvRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logging.WithContext(ph.ctx).
			Error("Could not unmarshal request", zap.Error(err))

		renderInvariantResponse(w, r,
			models.NewInvRequestUnmarshalErrResp())
		return
	}

	invUUID, err := ph.service.ProcessInvariantRequest(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.WithContext(ph.ctx).
			Error("Could not process invariant request", zap.Error(err))

		renderInvariantResponse(w, r, models.NewInvRequestUnmarshalErrResp())
		return
	}

	renderInvariantResponse(w, r, models.NewOkResp(invUUID))
}
