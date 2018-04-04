package util

import (
	"github.com/skycoin/skycoin/src/cipher"
	"sync"
)

type Addresses struct {
	mux  sync.Mutex
	dict map[cipher.Address]struct{}
}

func NewAddresses(count int) *Addresses {
	return &Addresses{
		dict: make(map[cipher.Address]struct{}, count),
	}
}

func (a *Addresses) AddPubKey(pk cipher.PubKey) {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.dict[cipher.AddressFromPubKey(pk)] = struct{}{}
}

func (a *Addresses) HasAddress(v cipher.Address) bool {
	a.mux.Lock()
	defer a.mux.Unlock()

	_, ok := a.dict[v]
	return ok
}
