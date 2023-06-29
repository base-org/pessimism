package core

type CtxKey uint8

const (
	Logger CtxKey = iota
	Metrics
	State
	L1Client
	L2Client
)
