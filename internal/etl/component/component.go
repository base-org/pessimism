package component

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

const (
	killSig = 0
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	/*
		NOTE - Storing the PUUID assumes that one component
		can only be a part of one pipeline at a time. This could be
		problematic if we want to have a component be a part of multiple
		pipelines at once. In that case, we would need to store a slice
		of PUUIDs instead.
	*/
	// PUUID ... Returns component's PipelineUUID
	PUUID() core.PipelineUUID
	// SetPUUID ... Sets component's PipelineUUID
	SetPUUID(pUUID core.PipelineUUID)

	// UUID ...
	UUID() core.ComponentUUID
	// Type ... Returns component enum type
	Type() core.ComponentType

	// AddRelay ... Adds an engine relay to component egress routing
	AddRelay(relay *core.EngineInputRelay) error

	// AddEgress ...
	AddEgress(core.ComponentUUID, chan core.TransitData) error
	// RemoveEgress ...
	RemoveEgress(core.ComponentUUID) error

	// Close ... Signifies a component to stop operating
	Close() error

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error

	// GetIngress ... Returns component ingress channel for some register type value
	GetIngress(rt core.RegisterType) (chan core.TransitData, error)

	// OutputType ... Returns component output data type
	OutputType() core.RegisterType

	StateKey() core.StateKey

	// TODO(#24): Add Internal Component Activity State Tracking
	ActivityState() ActivityState
}

// metaData ... Component agnostic struct that stores component metadata and routing state
type metaData struct {
	id    core.ComponentUUID
	pUUID core.PipelineUUID

	cType    core.ComponentType
	output   core.RegisterType
	state    ActivityState
	cacheKey core.StateKey
	inTypes  []core.RegisterType

	closeChan chan int
	stateChan chan StateChange

	*ingressHandler
	*egressHandler

	*sync.RWMutex
}

// newMetaData ... Initializer
func newMetaData(ct core.ComponentType, ot core.RegisterType) *metaData {
	return &metaData{
		id:    core.NilComponentUUID(),
		pUUID: core.NilPipelineUUID(),

		cType:          ct,
		egressHandler:  newEgressHandler(),
		ingressHandler: newIngressHandler(),
		state:          Inactive,
		closeChan:      make(chan int),
		stateChan:      make(chan StateChange),
		output:         ot,
		RWMutex:        &sync.RWMutex{},
	}
}

// ActivityState ... Returns component current activity state
func (meta *metaData) ActivityState() ActivityState {
	return meta.state
}

// StateKey ... Returns component's state key
func (meta *metaData) StateKey() core.StateKey {
	return meta.cacheKey
}

// UUID ... Returns component's ComponentUUID
func (meta *metaData) UUID() core.ComponentUUID {
	return meta.id
}
func (meta *metaData) SetPUUID(pUUID core.PipelineUUID) {
	meta.pUUID = pUUID
}

// UUID ... Returns component's PipelineUUID
// NOTE - This assumes that component collisions are impossible
func (meta *metaData) PUUID() core.PipelineUUID {
	return meta.pUUID
}

// Type ... Returns component's type
func (meta *metaData) Type() core.ComponentType {
	return meta.cType
}

// OutputType ... Returns component's data output type
func (meta *metaData) OutputType() core.RegisterType {
	return meta.output
}

// emitStateChange ... Emits a stateChange event to stateChan
func (meta *metaData) emitStateChange(as ActivityState) {
	event := StateChange{
		ID:   meta.id,
		From: meta.state,
		To:   as,
	}

	meta.state = as
	meta.stateChan <- event // Send to upstream consumers
}

// Option ... Component type agnostic option
type Option = func(*metaData)

// WithCUUID ... Passes component UUID to component metadata field
func WithCUUID(id core.ComponentUUID) Option {
	return func(meta *metaData) {
		meta.id = id
	}
}

// WithEventChan ... Passes state channel to component metadata field
func WithEventChan(sc chan StateChange) Option {
	return func(md *metaData) {
		md.stateChan = sc
	}
}

// WithInTypes	... Passes input types to component metadata field
func WithInTypes(its []core.RegisterType) Option {
	return func(md *metaData) {
		md.inTypes = its
	}
}

// WithStateKey ... Passes state key to component metadata field
func WithStateKey(key core.StateKey) Option {
	return func(md *metaData) {
		md.cacheKey = key
	}
}
