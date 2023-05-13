package models

import (
	"net/http"

	"github.com/base-org/pessimism/internal/core"
)

func NewOkResp(id core.InvSessionUUID) *InvResponse {
	return &InvResponse{
		Status: OK,
		Result: map[string]string{"invariant_uuid": id.String()},
	}
}

// NewInvRequestUnmarshalErrResp ... New unmarshal error response construction
func NewInvRequestUnmarshalErrResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Code:   http.StatusBadRequest,
		Error:  "could not unmarshal request body",
	}
}

// NewNoProcessInvErrResp ... New internal processing response error
func NewNoProcessInvErrResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Code:   http.StatusInternalServerError,
		Error:  "error processing invariant request",
	}
}
