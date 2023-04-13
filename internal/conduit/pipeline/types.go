package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/config"
)

// OracleType ...
type OracleType = string

const (
	// BackTestOracle ... Represents an oracle used for backtesting some invariant
	BacktestOracle OracleType = "backtest"
	// LiveOracle ... Represents an oracle used for powering some live invariant
	LiveOracle OracleType = "live"
)

// OutputRouter specific errors
const (
	dirAlreadyExistsErr = "%d directive key already exists within component router mapping"
	dirNotFoundErr      = "no directive key %d exists within component router mapping"
)

// Generalized component constructor types
type (
	// OracleConstructor ... Type declaration that a registry oracle component constructor must adhere to
	OracleConstructor = func(ctx context.Context, ot OracleType, cfg *config.OracleConfig,
		clientNew EthClientInterface) (Component, error)

	// PipeConstructorFunc ... Type declaration that a registry pipe component constructor must adhere to
	PipeConstructorFunc = func(ctx context.Context, inputChan chan models.TransitData) (Component, error)
)
