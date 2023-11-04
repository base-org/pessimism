package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
)

type AddressMap struct {
	m map[common.Address]map[core.PathID][]core.UUID
}

func NewAddressMap() *AddressMap {
	return &AddressMap{
		m: make(map[common.Address]map[core.PathID][]core.UUID),
	}
}

func (am *AddressMap) Get(address common.Address, PathID core.PathID) ([]core.UUID, error) {
	if _, found := am.m[address]; !found {
		return []core.UUID{}, fmt.Errorf("address provided is not tracked %s", address.String())
	}

	if _, found := am.m[address][PathID]; !found {
		return []core.UUID{}, fmt.Errorf("PathID provided is not tracked %s", PathID.String())
	}

	return am.m[address][PathID], nil
}

func (am *AddressMap) Insert(addr common.Address, PathID core.PathID, sUUID core.UUID) error {
	// 1. Check if address exists; create nested entry & return if not
	if _, found := am.m[addr]; !found {
		am.m[addr] = make(map[core.PathID][]core.UUID)
		am.m[addr][PathID] = []core.UUID{sUUID}
		return nil
	}

	// 2. Check if pipeline UUID exists; create entry & return if not
	if _, found := am.m[addr][PathID]; !found {
		am.m[addr][PathID] = []core.UUID{sUUID}
		return nil
	}

	// 3. Ensure that entry doesn't already exist
	for _, entry := range am.m[addr][PathID] {
		if entry == sUUID {
			return fmt.Errorf("entry already exists")
		}
	}

	// 4. Append entry and return
	am.m[addr][PathID] = append(am.m[addr][PathID], sUUID)
	return nil
}
