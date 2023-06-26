package models

import (
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/core"
)

// InvariantMethod ... Represents the invariant operation method
type InvariantMethod int

const (
	Run InvariantMethod = iota
	// NOTE - Update is not implemented yet
	Update
	// NOTE - Stop is not implemented yet
	Stop
)

func StringToInvariantMethod(s string) InvariantMethod {
	switch s {
	case "run":
		return Run
	case "update":
		return Update
	case "stop":
		return Stop
	default:
		return Run
	}
}

// InvResponseStatus ... Represents the invariant operation response status
type InvResponseStatus string

const (
	OK    InvResponseStatus = "OK"
	NotOK InvResponseStatus = "NOTOK"
)

// InvRequestParams ... Request params for invariant operation
type InvRequestParams struct {
	Network string `json:"network"`
	PType   string `json:"pipeline_type"`
	InvType string `json:"type"`

	StartHeight *big.Int `json:"start_height"`
	EndHeight   *big.Int `json:"end_height"`

	SessionParams map[string]interface{} `json:"invariant_params"`
	// TODO(#81): No Support for Multiple Alerting Destinations for an Invariant Session
	AlertingDest string `json:"alert_destination"`
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

// GeneratePipelineConfig ... Generates a pipeline config using the request params
func (irp *InvRequestParams) GeneratePipelineConfig(pollInterval time.Duration,
	regType core.RegisterType) *core.PipelineConfig {
	return &core.PipelineConfig{
		Network:      irp.NetworkType(),
		DataType:     regType,
		PipelineType: irp.PiplineType(),
		ClientConfig: &core.ClientConfig{
			Network:      irp.NetworkType(),
			PollInterval: pollInterval,
			StartHeight:  irp.StartHeight,
			EndHeight:    irp.EndHeight,
		},
	}
}

// SessionConfig ... Generates a session config using the request params
func (irp *InvRequestParams) SessionConfig() *core.SessionConfig {
	return &core.SessionConfig{
		AlertDest: irp.AlertingDestType(),
		Type:      irp.InvariantType(),
		Params:    irp.SessionParams,
	}
}

// InvRequestBody ... Request body for invariant operation request
type InvRequestBody struct {
	Method string           `json:"method"`
	Params InvRequestParams `json:"params"`
}

// MethodType ... Returns the invariant method type
func (irb *InvRequestBody) MethodType() InvariantMethod {
	return StringToInvariantMethod(irb.Method)
}

// InvResult ... Result of invariant operation
type InvResult = map[string]string

// InvResponse ... Response for invariant operation request
type InvResponse struct {
	Code   int               `json:"status_code"`
	Status InvResponseStatus `json:"status"`

	Result InvResult `json:"result"`
	Error  string    `json:"error"`
}
