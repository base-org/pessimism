package core

type CtxKey uint8

const (
	Logger CtxKey = iota
	State
	L1Client
	L2Client
)
