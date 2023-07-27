package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
)

// AddressingMap ... Interface for mapping addresses to session UUIDs
type AddressingMap interface {
	GetSUUIDsByPair(address common.Address, pUUID core.PUUID) ([]core.SUUID, error)
	Insert(addr common.Address, pUUID core.PUUID, sUUID core.SUUID) error
}

// addressingMap ... Implementation of AddressingMap
type addressingMap struct {
	m map[common.Address]map[core.PUUID][]core.SUUID
}

// GetSessionUUIDsByPair ... Gets the session UUIDs by the pair of address and pipeline UUID
func (am *addressingMap) GetSUUIDsByPair(address common.Address, pUUID core.PUUID) ([]core.SUUID, error) {
	if _, found := am.m[address]; !found {
		return []core.SUUID{}, fmt.Errorf("address provided is not tracked %s", address.String())
	}

	if _, found := am.m[address][pUUID]; !found {
		return []core.SUUID{}, fmt.Errorf("PUUID provided is not tracked %s", pUUID.String())
	}

	return am.m[address][pUUID], nil
}

// Insert ... Inserts a new entry into the addressing map
func (am *addressingMap) Insert(addr common.Address, pUUID core.PUUID, sUUID core.SUUID) error {
	// 1. Check if address exists; create nested entry & return if not
	if _, found := am.m[addr]; !found {
		am.m[addr] = make(map[core.PUUID][]core.SUUID)
		am.m[addr][pUUID] = []core.SUUID{sUUID}
		return nil
	}

	// 2. Check if pipeline UUID exists; create entry & return if not
	if _, found := am.m[addr][pUUID]; !found {
		am.m[addr][pUUID] = []core.SUUID{sUUID}
		return nil
	}

	// 3. Ensure that entry doesn't already exist
	for _, entry := range am.m[addr][pUUID] {
		if entry == sUUID {
			return fmt.Errorf("entry already exists")
		}
	}

	// 4. Append entry and return
	am.m[addr][pUUID] = append(am.m[addr][pUUID], sUUID)
	return nil
}

// NewAddressingMap ... Initializer
func NewAddressingMap() AddressingMap {
	return &addressingMap{
		m: make(map[common.Address]map[core.PUUID][]core.SUUID),
	}
}
