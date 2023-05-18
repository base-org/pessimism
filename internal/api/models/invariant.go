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
	Network string `json:"network"`
	PType   string `json:"pipeline_type"`
	InvType string `json:"type"`

	StartHeight *big.Int `json:"start_height"`
	EndHeight   *big.Int `json:"end_height"`

	SessionParams map[string]interface{} `json:"invariant_params"`
	AlertingDest  string                 `json:"alert_destination"`
}

// AlertingDestType ... Returns the alerting destination type
func (irp *InvRequestParams) AlertingDestType() core.AlertDestination {
	return core.StringToAlertingDestType(irp.AlertingDest)
}

// NetworkType ... Returns the network type
func (irp *InvRequestParams) NetworkType() core.Network {
	return core.StringToNetwork(irp.Network)
}

// PiplineType ... Returns the pipeline type
func (irp *InvRequestParams) PiplineType() core.PipelineType {
	return core.StringToPipelineType(irp.PType)
}

// InvariantType ... Returns the invariant type
func (irp *InvRequestParams) InvariantType() core.InvariantType {
	return core.StringToInvariantType(irp.InvType)
}

type InvRequestBody struct {
	Method InvariantMethod  `json:"method"`
	Params InvRequestParams `json:"params"`
}

type InvResult = map[string]string

type InvResponse struct {
	Code   int               `json:"status_code"`
	Status InvResponseStatus `json:"status"`

	Result InvResult `json:"result"`
	Error  string    `json:"error"`
}

// GeneratePipelineConfig ... Generates a pipeline config using the request params
func (params *InvRequestParams) GeneratePipelineConfig(endpoint string, pollInterval time.Duration,
	regType core.RegisterType) *core.PipelineConfig {
	return &core.PipelineConfig{
		Network:      params.NetworkType(),
		DataType:     regType,
		PipelineType: params.PiplineType(),
		OracleCfg: &core.OracleConfig{
			RPCEndpoint:  endpoint,
			PollInterval: pollInterval,
			StartHeight:  params.StartHeight,
			EndHeight:    params.EndHeight,
		},
	}
}
