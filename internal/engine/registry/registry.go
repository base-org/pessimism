package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

// GetInvariant ... Returns an invariant based on the invariant type provided
// a general config
func GetInvariant(ctx context.Context, it core.InvariantType, cfg any) (invariant.Invariant, error) {
	var inv invariant.Invariant

	switch it {
	case core.BalanceEnforcement:
		cfg, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		// convert json to struct
		invConfg := BalanceInvConfig{}
		err = json.Unmarshal(cfg, &invConfg)
		if err != nil {
			return nil, err
		}

		inv = NewBalanceInvariant(&invConfg)

	case core.ContractEvent:
		cfg, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		// convert json to struct
		invConfg := EventInvConfig{}
		err = json.Unmarshal(cfg, &invConfg)
		if err != nil {
			return nil, err
		}

		inv = NewEventInvariant(&invConfg)

	case core.WithdrawalEnforcement:
		cfg, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		// convert json to struct
		invConfg := WthdrawlEnforceCfg{}
		err = json.Unmarshal(cfg, &invConfg)
		if err != nil {
			return nil, err
		}

		return NewWthdrawlEnforceInv(ctx, &invConfg)

	default:
		return nil, fmt.Errorf("could not find implementation for type %s", it.String())
	}

	return inv, nil
}
