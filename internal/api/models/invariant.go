package models

import (
	"math/big"
	"time"

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

	SessionParams map[string]interface{} `json:"invariant_params"`
}

type InvRequestBody struct {
	Method InvariantMethod  `json:"method"`
	Params InvRequestParams `json:"params"`
}

type InvResponse struct {
	Code   int               `json:"status_code"`
	Status InvResponseStatus `json:"status"`

	Result InvResult `json:"result"`
	Error  string    `json:"error"`
}

type InvResult = map[string]string

func (params *InvRequestParams) GeneratePipelineConfig(endpoint string, pollInterval time.Duration,
	regType core.RegisterType) *core.PipelineConfig {
	return &core.PipelineConfig{
		Network:      params.Network,
		DataType:     regType,
		PipelineType: params.PType,
		OracleCfg: &core.OracleConfig{
			RPCEndpoint:  endpoint,
			PollInterval: pollInterval,
			StartHeight:  params.StartHeight,
			EndHeight:    params.EndHeight,
		},
	}
}
