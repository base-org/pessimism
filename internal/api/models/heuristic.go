package models

import (
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/core"
)

// HeuristicMethod ... Represents the heuristic operation method
type HeuristicMethod int

const (
	Run HeuristicMethod = iota
	// NOTE - Update is not implemented yet
	Update
	// NOTE - Stop is not implemented yet
	Stop
)

func StringToHeuristicMethod(s string) HeuristicMethod {
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

// InvResponseStatus ... Represents the heuristic operation response status
type InvResponseStatus string

const (
	OK    InvResponseStatus = "OK"
	NotOK InvResponseStatus = "NOTOK"
)

// InvRequestParams ... Request params for heuristic operation
type InvRequestParams struct {
	Network string `json:"network"`
	PType   string `json:"pipeline_type"`
	InvType string `json:"type"`

	StartHeight *big.Int `json:"start_height"`
	EndHeight   *big.Int `json:"end_height"`

	SessionParams map[string]interface{} `json:"heuristic_params"`
	// TODO(#81): No Support for Multiple Alerting Destinations for an Heuristic Session
	AlertingDest string `json:"alert_destination"`
}

// Params ... Returns the heuristic session params
func (irp *InvRequestParams) Params() *core.InvSessionParams {
	isp := core.NewSessionParams()

	for k, v := range irp.SessionParams {
		isp.SetValue(k, v)
	}

	return isp
}

// AlertingDestType ... Returns the alerting destination type
func (irp *InvRequestParams) AlertingDestType() core.AlertDestination {
	return core.StringToAlertingDestType(irp.AlertingDest)
}

// NetworkType ... Returns the network type
func (irp *InvRequestParams) NetworkType() core.Network {
	return core.StringToNetwork(irp.Network)
}

// PipelineType ... Returns the pipeline type
func (irp *InvRequestParams) PipelineType() core.PipelineType {
	return core.StringToPipelineType(irp.PType)
}

// HeuristicType ... Returns the heuristic type
func (irp *InvRequestParams) HeuristicType() core.HeuristicType {
	return core.StringToHeuristicType(irp.InvType)
}

// GeneratePipelineConfig ... Generates a pipeline config using the request params
func (irp *InvRequestParams) GeneratePipelineConfig(pollInterval time.Duration,
	regType core.RegisterType) *core.PipelineConfig {
	return &core.PipelineConfig{
		Network:      irp.NetworkType(),
		DataType:     regType,
		PipelineType: irp.PipelineType(),
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
		Type:      irp.HeuristicType(),
		Params:    irp.Params(),
		PT:        irp.PipelineType(),
	}
}

// InvRequestBody ... Request body for heuristic operation request
type InvRequestBody struct {
	Method string           `json:"method"`
	Params InvRequestParams `json:"params"`
}

func (irb *InvRequestBody) Clone() *InvRequestBody {
	return &InvRequestBody{
		Method: irb.Method,
		Params: irb.Params,
	}
}

// MethodType ... Returns the heuristic method type
func (irb *InvRequestBody) MethodType() HeuristicMethod {
	return StringToHeuristicMethod(irb.Method)
}

// InvResult ... Result of heuristic operation
type InvResult = map[string]string

// InvResponse ... Response for heuristic operation request
type InvResponse struct {
	Code   int               `json:"status_code"`
	Status InvResponseStatus `json:"status"`

	Result InvResult `json:"result"`
	Error  string    `json:"error"`
}
