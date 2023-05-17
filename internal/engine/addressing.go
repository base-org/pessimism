package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
)

type AddressingMap interface {
	GetSessionUUIDByPair(address common.Address, pUUID core.PipelineUUID) (core.InvSessionUUID, error)
	Insert(pUUID core.PipelineUUID, sUUID core.InvSessionUUID, address common.Address) error
}

type addressEntry struct {
	address common.Address
	sUUID   core.InvSessionUUID
	pUUID   core.PipelineUUID
}

type addressingMap struct {
	m map[common.Address][]addressEntry
}

func (am *addressingMap) GetSessionUUIDByPair(address common.Address, pUUID core.PipelineUUID) (core.InvSessionUUID, error) {
	if _, found := am.m[address]; !found {
		return core.NilInvariantUUID(), fmt.Errorf("address provided is not tracked %s", address.String())
	}

	// Now we know it's entry has been seen
	for _, entry := range am.m[address] {
		if entry.pUUID == pUUID { // Found
			return entry.sUUID, nil
		}
	}

	return core.NilInvariantUUID(), fmt.Errorf("could not find matching pUUID %s", pUUID.String())
}

func (am *addressingMap) Insert(pUUID core.PipelineUUID,
	sUUID core.InvSessionUUID, address common.Address) error {
	newEntry := addressEntry{
		address: address,
		sUUID:   sUUID,
		pUUID:   pUUID}

	if _, found := am.m[address]; !found {
		am.m[address] = []addressEntry{newEntry}
		return nil
	}

	// Now we know it's entry has been seen
	for _, entry := range am.m[address] {
		if entry.pUUID == pUUID {
			return fmt.Errorf("%s already exists for suuid %s", address, sUUID.String())
		}
	}

	am.m[address] = append(am.m[address], newEntry)
	return nil
}

func NewAddressingMap() AddressingMap {
	return &addressingMap{
		m: make(map[common.Address][]addressEntry),
	}
}
