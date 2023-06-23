package core

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TransitOption ... Option used to initialize transit data
type TransitOption = func(*TransitData)

// WithAddress ... Injects address to transit data
func WithAddress(address common.Address) TransitOption {
	return func(td *TransitData) {
		td.Address = address
	}
}

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	Network Network
	Type    RegisterType

	Address common.Address
	Value   any
}

// NewTransitData ... Initializes transit data with supplied options
// NOTE - transit data is used as a standard data representation
// for commmunication between all ETL components and the risk engine
func NewTransitData(rt RegisterType, val any, opts ...TransitOption) TransitData {
	td := TransitData{
		Timestamp: time.Now(),
		Type:      rt,
		Value:     val,
	}

	for _, opt := range opts { // Apply options
		opt(&td)
	}

	return td
}

// Addressed ... Indicates whether the transit data has an
// associated address field
func (td *TransitData) Addressed() bool {
	return td.Address != common.Address{0}
}

// NewTransitChannel ... Builds new tranit channel
func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}

// InvariantInput ... Standardized type used to supply
// the Risk Engine
type InvariantInput struct {
	PUUID PUUID
	Input TransitData
}

// EngineInputRelay ... Represents a inter-subsystem
// relay used to bind final ETL pipeline outputs to risk engine inputs
type EngineInputRelay struct {
	pUUID   PUUID
	outChan chan InvariantInput
}

// NewEngineRelay ... Initializer
func NewEngineRelay(pUUID PUUID, outChan chan InvariantInput) *EngineInputRelay {
	return &EngineInputRelay{
		pUUID:   pUUID,
		outChan: outChan,
	}
}

// RelayTransitData ... Creates invariant input from transit data to send to risk engine
func (eir *EngineInputRelay) RelayTransitData(td TransitData) error {
	invInput := InvariantInput{
		PUUID: eir.pUUID,
		Input: td,
	}

	eir.outChan <- invInput
	return nil
}

const (
	AddressKey = "address"
	NestedArgs = "args"
)

// InvSessionParams ... Parameters used to initialize an invariant session
type InvSessionParams map[string]interface{}

// Address ... Returns the address from the invariant session params
func (sp *InvSessionParams) Address() string {
	rawAddr, found := (*sp)[AddressKey]
	if !found {
		return ""
	}

	addr, success := rawAddr.(string)
	if !success {
		return ""
	}

	return addr
}

// Address ... Returns the address from the invariant session params
func (sp *InvSessionParams) NestedArgs() []string {
	rawArgs, found := (*sp)[NestedArgs]
	if !found {
		return []string{}
	}

	args, success := rawArgs.([]interface{})
	if !success {
		return []string{}
	}

	var strArgs []string
	for _, arg := range args {
		strArgs = append(strArgs, arg.(string))
	}

	return strArgs
}

// InvalOutcome ... Represents an invalidation outcome
type InvalOutcome struct {
	TimeStamp time.Time
	Message   string
}

// Subsystem ... Represents a subsystem
type Subsystem interface {
	EventLoop() error
	Shutdown() error
}
