package models

// ComponentType
type ComponentType int

const (
	Oracle   ComponentType = 0
	Pipe     ComponentType = 1
	Conveyor ComponentType = 2
)

type Timeouts int

const (
	EthClientTimeout Timeouts = 20 // in seconds
)
