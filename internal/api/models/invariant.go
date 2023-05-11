package models

import (
	"math/big"

	"github.com/base-org/pessimism/internal/core"
)

type InvariantMethod int

const (
	Run InvariantMethod = iota
)

type InvResponseStatus string

const (
	OK    InvResponseStatus = "OK"
	NotOK InvResponseStatus = "NOTOK"
)

type InvRequestParams struct {
	Network core.Network       `json:"network"`
	PType   core.PipelineType  `json:"pipeline_type"`
	InvType core.InvariantType `json:"type"`

	StartHeight *big.Int `json:"start_height"`
	EndHeight   *big.Int `json:"end_height"`

	SessionParams interface{} `json:"invariant_params"`
}

type InvRequestBody struct {
	Method InvariantMethod  `json:"method"`
	Params InvRequestParams `json:"params"`
}

type InvResponse struct {
	Status InvResponseStatus `json:"status"`

	Result any    `json:"result"`
	Error  string `json:"error"`
}

func NewOkResp(id core.InvSessionUUID) *InvResponse {
	return &InvResponse{
		Status: OK,
		Result: map[string]string{"invariant_uuid": string(id.String())},
	}
}

// NewInvRequestUnmarshalErrResp ... New unmarshal error response construction
func NewInvRequestUnmarshalErrResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Error:  "could not unmarshal request body",
	}
}

// NewNoProcessInvErrResp ...
func NewNoProcessInvErrResp() *InvResponse {
	return &InvResponse{
		Status: NotOK,
		Error:  "error processing invariant request",
	}
}
