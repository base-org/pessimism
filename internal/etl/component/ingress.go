package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type ingress struct {
	entryPoints map[core.RegisterType]chan core.TransitData
}

func newIngress() *ingress {
	return &ingress{
		entryPoints: make(map[core.RegisterType]chan core.TransitData),
	}
}

func (in *ingress) GetEntryPoint(rt core.RegisterType) (chan core.TransitData, error) {
	val, found := in.entryPoints[rt]
	if !found {
		return nil, fmt.Errorf(entryNotFoundErr, rt.String())
	}

	return val, nil
}

func (in *ingress) createEntryPoint(rt core.RegisterType) error {
	if _, found := in.entryPoints[rt]; found {
		return fmt.Errorf(entryAlreadyExistsErr, rt.String())
	}

	in.entryPoints[rt] = make(chan core.TransitData)

	return nil
}
