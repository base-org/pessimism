package core

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type AccountBalanceVal struct {
	Address common.Address
	Balance *big.Int
}

// TransitData ... Standardized type used for data inter-communication
// between all ETL components and Risk Engine
type TransitData struct {
	Timestamp time.Time

	Type  RegisterType
	Value any
}

// NewTransitChannel ... Builds new tranit channel
func NewTransitChannel() chan TransitData {
	return make(chan TransitData)
}
