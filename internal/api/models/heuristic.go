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

// SessionResponseStatus ... Represents the heuristic operation response status
type SessionResponseStatus string

const (
	OK    SessionResponseStatus = "OK"
	NotOK SessionResponseStatus = "NOTOK"
)

// SessionRequestParams ... Request params for heuristic operation
type SessionRequestParams struct {
	Network       string `json:"network"`
	HeuristicType string `json:"type"`

	StartHeight *big.Int `json:"start_height"`
	EndHeight   *big.Int `json:"end_height"`

	SessionParams  map[string]interface{} `json:"heuristic_params"`
	AlertingParams *core.AlertPolicy      `json:"alerting_params"`
}

// Params ... Returns the heuristic session params
func (hrp *SessionRequestParams) Params() *core.SessionParams {
	isp := core.NewSessionParams(hrp.NetworkType())

	for k, v := range hrp.SessionParams {
		isp.SetValue(k, v)
	}

	return isp
}

// AlertingDestType ... Returns the alerting destination type
func (hrp *SessionRequestParams) AlertingDestType() core.AlertDestination {
	return hrp.AlertingParams.Destination()
}

// NetworkType ... Returns the network type
func (hrp *SessionRequestParams) NetworkType() core.Network {
	return core.StringToNetwork(hrp.Network)
}

// Heuristic ... Returns the heuristic type
func (hrp *SessionRequestParams) Heuristic() core.HeuristicType {
	return core.StringToHeuristicType(hrp.HeuristicType)
}

func (hrp *SessionRequestParams) AlertPolicy() *core.AlertPolicy {
	return hrp.AlertingParams
}

// NewPathCfg ... Generates a path config using the request params
func (hrp *SessionRequestParams) NewPathCfg(pollInterval time.Duration,
	regType core.TopicType) *core.PathConfig {
	return &core.PathConfig{
		Network:  hrp.NetworkType(),
		DataType: regType,
		PathType: core.Live,
		ClientConfig: &core.ClientConfig{
			Network:      hrp.NetworkType(),
			PollInterval: pollInterval,
			StartHeight:  hrp.StartHeight,
			EndHeight:    hrp.EndHeight,
		},
	}
}

// SessionConfig ... Generates a session config using the request params
func (hrp *SessionRequestParams) SessionConfig() *core.SessionConfig {
	return &core.SessionConfig{
		AlertPolicy: hrp.AlertPolicy(),
		Type:        hrp.Heuristic(),
		Params:      hrp.Params(),
		PT:          core.Live,
	}
}

// SessionRequestBody ... Request body for heuristic operation request
type SessionRequestBody struct {
	Method string               `json:"method"`
	Params SessionRequestParams `json:"params"`
}

func (irb *SessionRequestBody) Clone() *SessionRequestBody {
	return &SessionRequestBody{
		Method: irb.Method,
		Params: irb.Params,
	}
}

// MethodType ... Returns the heuristic method type
func (irb *SessionRequestBody) MethodType() HeuristicMethod {
	return StringToHeuristicMethod(irb.Method)
}

// Result ... Result of heuristic operation
type Result = map[string]string

// SessionResponse ... Response for heuristic operation request
type SessionResponse struct {
	Code   int                   `json:"status_code"`
	Status SessionResponseStatus `json:"status"`

	Result Result `json:"result"`
	Error  string `json:"error"`
}
