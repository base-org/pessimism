package models

// ComponentType
type ComponentType int

const (
	Oracle   ComponentType = 0
	Pipe     ComponentType = 1
	Conveyor ComponentType = 2
)

type FetchType int

const (
	FetchHeader FetchType = 0
	FetchBlock  FetchType = 1
)

type Timeouts int

const (
	EthClientTimeout Timeouts = 20 // in seconds
)
