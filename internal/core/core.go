package core

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"

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

// WithOriginTS ... Injects origin timestamp to transit data
func WithOriginTS(t time.Time) TransitOption {
	return func(td *TransitData) {
		td.OriginTS = t
	}
}

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	OriginTS  time.Time
	Timestamp time.Time

	Network Network
	Type    RegisterType

	Address common.Address
	Value   any
}

// NewTransitData ... Initializes transit data with supplied options
// NOTE - transit data is used as a standard data representation
// for communication between all ETL components and the risk engine
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

// NewTransitChannel ... Builds new transit channel
func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}

// HeuristicInput ... Standardized type used to supply
// the Risk Engine
type HeuristicInput struct {
	PUUID PUUID
	Input TransitData
}

// ExecInputRelay ... Represents a inter-subsystem
// relay used to bind final ETL pipeline outputs to risk engine inputs
type ExecInputRelay struct {
	pUUID   PUUID
	outChan chan HeuristicInput
}

// NewEngineRelay ... Initializer
func NewEngineRelay(pUUID PUUID, outChan chan HeuristicInput) *ExecInputRelay {
	return &ExecInputRelay{
		pUUID:   pUUID,
		outChan: outChan,
	}
}

// RelayTransitData ... Creates heuristic input from transit data to send to risk engine
func (eir *ExecInputRelay) RelayTransitData(td TransitData) error {
	hi := HeuristicInput{
		PUUID: eir.pUUID,
		Input: td,
	}

	eir.outChan <- hi
	return nil
}

const (
	AddressKey = "address"
	NestedArgs = "args"
)

// SessionParams ... Parameters used to initialize a heuristic session
type SessionParams struct {
	params map[string]any
}

// Bytes ... Returns a marshalled byte array of the heuristic session params
func (sp *SessionParams) Bytes() []byte {
	bytes, _ := json.Marshal(sp.params)
	return bytes
}

// NewSessionParams ... Initializes heuristic session params
func NewSessionParams() *SessionParams {
	isp := &SessionParams{
		params: make(map[string]any, 0),
	}
	isp.params[NestedArgs] = []any{}
	return isp
}

func (sp *SessionParams) Value(key string) (any, error) {
	val, found := sp.params[key]
	if !found {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return val, nil
}

// Address ... Returns the address from the heuristic session params
func (sp *SessionParams) Address() common.Address {
	rawAddr, found := sp.params[AddressKey]
	if !found {
		return common.Address{0}
	}

	addr, success := rawAddr.(string)
	if !success {
		return common.Address{0}
	}

	return common.HexToAddress(addr)
}

// SetValue ... Sets a value in the heuristic session params
func (sp *SessionParams) SetValue(key string, val any) {
	sp.params[key] = val
}

// SetNestedArg ... Sets a nested argument in the heuristic session params
// unique nested key/value space
func (sp *SessionParams) SetNestedArg(arg interface{}) {
	args := sp.NestedArgs()
	args = append(args, arg)
	sp.params[NestedArgs] = args
}

// NestedArgs ... Returns the nested arguments from the heuristic session params
func (sp *SessionParams) NestedArgs() []any {
	rawArgs, found := sp.params[NestedArgs]
	if !found {
		return []any{}
	}

	args, success := rawArgs.([]any)
	if !success {
		return []any{}
	}

	return args
}

// Subsystem ... Represents a subsystem
type Subsystem interface {
	EventLoop() error
	Shutdown() error
}

const (
	L1Portal            = "l1_portal_address" //#nosec G101: False positive, this isn't a credential
	L2ToL1MessagePasser = "l2_to_l1_address"  //#nosec G101: False positive, this isn't a credential
	L2OutputOracle      = "l2_output_address" //#nosec G101: False positive, this isn't a credential
)

// Regexp for parsing yaml files
var reVar = regexp.MustCompile(`^\${(\w+)}$`)

type StringFromEnv string

// UnmarshalYAML implements the yaml.Unmarshaler interface to allow parsing strings from env vars.
func (e *StringFromEnv) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	if match := reVar.FindStringSubmatch(s); len(match) > 0 {
		*e = StringFromEnv(os.Getenv(match[1]))
	} else {
		*e = StringFromEnv(s)
	}
	return nil
}

// String returns the string value, implementing the flag.Value interface.
func (e *StringFromEnv) String() string {
	return string(*e)
}
