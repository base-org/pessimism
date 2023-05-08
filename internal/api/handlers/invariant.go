package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// RunInvariant ... Handle invariant run request
func (ph *PessimismHandler) RunInvariant(w http.ResponseWriter, r *http.Request) {
	var body models.InvRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logging.NoContext().Error("could not unmarshal request", zap.Error(err))

		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, models.NewUnmarshalErrResp())
		return
	}

	uuid, err := ph.service.ProcessInvariantRequest(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.NoContext().Error("could not process invariant request", zap.Error(err))
		render.JSON(w, r, models.NewUnmarshalErrResp())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, models.NewOkResp(uuid))
}
