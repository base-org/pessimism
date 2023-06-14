package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
)

// AddressingMap ... Interface for mapping addresses to session UUIDs
type AddressingMap interface {
	GetSessionUUIDByPair(address common.Address, pUUID core.PUUID) (core.SUUID, error)
	Insert(pUUID core.PUUID, sUUID core.SUUID, address common.Address) error
}

// addressEntry ... Entry for the addressing map
type addressEntry struct {
	address common.Address
	sUUID   core.SUUID
	pUUID   core.PUUID
}

// addressingMap ... Implementation of AddressingMap
type addressingMap struct {
	m map[common.Address][]addressEntry
}

// GetSessionUUIDByPair ... Gets the session UUID by the pair of address and pipeline UUID
func (am *addressingMap) GetSessionUUIDByPair(address common.Address,
	pUUID core.PUUID) (core.SUUID, error) {
	if _, found := am.m[address]; !found {
		return core.NilSUUID(), fmt.Errorf("address provided is not tracked %s", address.String())
	}

	// Now we know it's entry has been seen
	for _, entry := range am.m[address] {
		if entry.pUUID == pUUID { // Found
			return entry.sUUID, nil
		}
	}

	return core.NilSUUID(), fmt.Errorf("could not find matching pUUID %s", pUUID.String())
}

// Insert ... Inserts a new entry into the addressing map
func (am *addressingMap) Insert(pUUID core.PUUID,
	sUUID core.SUUID, address common.Address) error {
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

// NewAddressingMap ... Initializer
func NewAddressingMap() AddressingMap {
	return &addressingMap{
		m: make(map[common.Address][]addressEntry),
	}
}
