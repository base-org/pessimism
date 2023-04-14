package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type ingressHandler struct {
	ingreses map[core.RegisterType]chan core.TransitData
}

func newIngressHandler() *ingressHandler {
	return &ingressHandler{
		ingreses: make(map[core.RegisterType]chan core.TransitData),
	}
}

func (ih *ingressHandler) GetIngress(rt core.RegisterType) (chan core.TransitData, error) {
	val, found := ih.ingreses[rt]
	if !found {
		return nil, fmt.Errorf(ingressNotFoundErr, rt.String())
	}

	return val, nil
}

func (ih *ingressHandler) createIngress(rt core.RegisterType) error {
	if _, found := ih.ingreses[rt]; found {
		return fmt.Errorf(ingressAlreadyExistsErr, rt.String())
	}

	ih.ingreses[rt] = core.NewTransitChannel()

	return nil
}
