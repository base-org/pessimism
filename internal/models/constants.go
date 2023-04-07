package models

// ComponentType
type ComponentType string

const (
	Oracle     ComponentType = "Oracle"
	Pipe       ComponentType = "Pipe"
	Aggregator ComponentType = "Aggregator"
)

type PipelineType = string

const (
	Backtest PipelineType = "Backtest"
	Live     PipelineType = "Live"
	MockTest PipelineType = "Mocktest"
)
