package core

import (
	"math/big"
)

// OracleConfig ... Configuration passed through to an oracle component constructor
type OracleConfig struct {
	RPCEndpoint  string
	StartHeight  *big.Int
	EndHeight    *big.Int
	NumOfRetries int
}

// PipelineConfig ... Configuration passed through to a pipeline constructor
type PipelineConfig struct {
	Network      Network
	DataType     RegisterType
	PipelineType PipelineType
	OracleCfg    *OracleConfig
}

func (oc *OracleConfig) Backfill() bool {
	return oc.StartHeight != nil
}

func (oc *OracleConfig) Backtest() bool {
	return oc.EndHeight != nil
}
