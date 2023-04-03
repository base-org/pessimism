package models

// ComponentType
type ComponentType int

const (
	Oracle   ComponentType = 0
	Pipe     ComponentType = 1
	Conveyor ComponentType = 2
)

type PipelineType = int

const (
	Backtest PipelineType = 0
	LiveTest PipelineType = 1
	MockTest PipelineType = 2
)
