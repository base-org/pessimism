package registry

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

func GetInvariant(it core.InvariantType, cfg any) (invariant.Invariant, error) {
	log.Printf("cfg %+v", cfg)

	switch it {
	case core.ExampleInv:

		cfg, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		// convert json to struct
		invConfg := ExampleInvConfig{}
		err = json.Unmarshal(cfg, &invConfg)
		if err != nil {
			return nil, err
		}

		inv := NewExampleInvariant(&invConfg)

		return inv, nil

	case core.TxCaller:

		cfg, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		// convert json to struct
		invConfg := InvocationInvConfig{}
		err = json.Unmarshal(cfg, &invConfg)
		if err != nil {
			return nil, err
		}

		inv := NewInvocationTrackerInvariant(&invConfg)

		return inv, nil

	default:
		return nil, fmt.Errorf("could not find implementation for type %s", it.String())
	}
}
