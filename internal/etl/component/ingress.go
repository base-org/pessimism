package component

import (
	"fmt"
	"log"

	"github.com/base-org/pessimism/internal/models"
)

type ingress struct {
	entryPoints map[models.RegisterType]chan models.TransitData
}

func newIngress() *ingress {
	return &ingress{
		entryPoints: make(map[models.RegisterType]chan models.TransitData),
	}
}

func (in *ingress) GetEntryPoint(rt models.RegisterType) (chan models.TransitData, error) {
	log.Printf("Fetching entrypoint for %s", rt)
	val, found := in.entryPoints[rt]
	if !found {
		return nil, fmt.Errorf("Could not find entrypoint")
	}

	return val, nil
}

func (in *ingress) CreateEntryPoint(rt models.RegisterType) error {
	// TODO - Duplication check
	log.Printf("Adding entrypoint for %s", rt)
	in.entryPoints[rt] = make(chan models.TransitData)

	return nil
}
