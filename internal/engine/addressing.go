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

func (am *AddressMap) Get(address common.Address, id core.PathID) ([]core.UUID, error) {
	if _, found := am.m[address]; !found {
		return []core.UUID{}, fmt.Errorf("address provided is not tracked %s", address.String())
	}

	if _, found := am.m[address][id]; !found {
		return []core.UUID{}, fmt.Errorf("id provided is not tracked %s", id.String())
	}

	return am.m[address][id], nil
}

func (am *AddressMap) Insert(addr common.Address, id core.PathID, uuid core.UUID) error {
	// 1. Check if address exists; create nested entry & return if not
	if _, found := am.m[addr]; !found {
		am.m[addr] = make(map[core.PathID][]core.UUID)
		am.m[addr][id] = []core.UUID{uuid}
		return nil
	}

	// 2. Check if path UUID exists; create entry & return if not
	if _, found := am.m[addr][id]; !found {
		am.m[addr][id] = []core.UUID{uuid}
		return nil
	}

	// 3. Ensure that entry doesn't already exist
	for _, entry := range am.m[addr][id] {
		if entry == uuid {
			return fmt.Errorf("entry already exists")
		}
	}

	// 4. Append entry and return
	am.m[addr][id] = append(am.m[addr][id], uuid)
	return nil
}
