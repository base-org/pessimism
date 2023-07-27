package models

import (
	"net/http"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// NewSessionAcceptedResp ...Returns an heuristic response with status accepted
func NewSessionAcceptedResp(id core.SUUID) *SessionResponse {
	return &SessionResponse{
		Status: OK,
		Code:   http.StatusAccepted,
		Result: Result{logging.SUUIDKey: id.String()},
	}
}

// NewSessionUnmarshalErrResp ... New unmarshal error response construction
func NewSessionUnmarshalErrResp() *SessionResponse {
	return &SessionResponse{
		Status: NotOK,
		Code:   http.StatusBadRequest,
		Error:  "could not unmarshal request body",
	}
}

// NewSessionNoProcessResp ... New internal processing response error
func NewSessionNoProcessResp() *SessionResponse {
	return &SessionResponse{
		Status: NotOK,
		Code:   http.StatusInternalServerError,
		Error:  "error processing heuristic request",
	}
}
