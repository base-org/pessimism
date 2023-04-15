package component

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// ingressHandler ... Used to manage ingresses for some component
// NOTE:  An edge is only possible between two components (C0, C1) where C0 -> C1
// if C0.outGresses[registerType] ⊆ C1.ingresses
type ingressHandler struct {
	ingreses map[core.RegisterType]chan core.TransitData
}

// newIngressHandler ... Initializer
func newIngressHandler() *ingressHandler {
	return &ingressHandler{
		ingreses: make(map[core.RegisterType]chan core.TransitData),
	}
}

// GetIngress ... Fetches ingress channel for some register type
func (ih *ingressHandler) GetIngress(rt core.RegisterType) (chan core.TransitData, error) {
	val, found := ih.ingreses[rt]
	if !found {
		return nil, fmt.Errorf(ingressNotFoundErr, rt.String())
	}

	return val, nil
}

// createIngress ... Creates ingress channel for some register type
func (ih *ingressHandler) createIngress(rt core.RegisterType) error {
	if _, found := ih.ingreses[rt]; found {
		return fmt.Errorf(ingressAlreadyExistsErr, rt.String())
	}

	ih.ingreses[rt] = core.NewTransitChannel()

	return nil
}
