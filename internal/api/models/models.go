package models

import (
	"net/http"

	"github.com/base-org/pessimism/internal/core"
)

// NewInvAcceptedResp ...Returns an invariant response with status accepted
func NewInvAcceptedResp(id core.SUUID) *InvResponse {
	return &InvResponse{
		Status: OK,
		Code:   http.StatusAccepted,
		Result: InvResult{core.SUUIDKey: id.String()},
	}
}

// NewInvUnmarshalErrResp ... New unmarshal error response construction
func NewInvUnmarshalErrResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Code:   http.StatusBadRequest,
		Error:  "could not unmarshal request body",
	}
}

// NewInvNoProcessInvResp ... New internal processing response error
func NewInvNoProcessInvResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Code:   http.StatusInternalServerError,
		Error:  "error processing invariant request",
	}
}
