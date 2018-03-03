package iko

import (
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"sync"
)

// StateDB records the state of the blockchain.
type StateDB interface {

	// GetKittyState obtains the current state of a kitty.
	// This consists of:
	//		- The address that the kitty resides under.
	//		- Transactions associated with the kitty.
	// It should return false if kitty of specified ID does not exist.
	GetKittyState(kittyID KittyID) (*KittyState, bool)

	// GetAddressState obtains the current state of an address.
	// This consists of:
	//		- Kitties owned by the address.
	//		- Transactions associated with the address.
	// The array of kitty IDs should be in ascending sequential order, from smallest index to highest.
	GetAddressState(address cipher.Address) *AddressState

	// AddKitty adds a kitty to the state under the specified address.
	// This should fail if:
	// 		- kitty of specified ID already exists in state.
	AddKitty(tx TxHash, kittyID KittyID, address cipher.Address) error

	// MoveKitty moves a kitty from one address to another.
	// This should fail if:
	//		- kitty of specified ID already belongs to the address ('from' and 'to' addresses are the same).
	//		- kitty of specified ID does not exist.
	//		- kitty of specified ID does not originally belong to the 'from' address.
	MoveKitty(tx TxHash, kittyID KittyID, from, to cipher.Address) error
}

type MemoryState struct {
	sync.Mutex
	kitties   map[KittyID]*KittyState
	addresses map[cipher.Address]*AddressState
}

func NewMemoryState() *MemoryState {
	return &MemoryState{
		kitties:   make(map[KittyID]*KittyState),
		addresses: make(map[cipher.Address]*AddressState),
	}
}

func (s *MemoryState) GetKittyState(kittyID KittyID) (*KittyState, bool) {
	s.Lock()
	defer s.Unlock()

	kState, ok := s.kitties[kittyID]
	return kState, ok
}

func (s MemoryState) GetAddressState(address cipher.Address) *AddressState {
	s.Lock()
	defer s.Unlock()

	aState, ok := s.addresses[address]
	if !ok {
		aState = NewAddressState()
	}
	return aState
}

func (s *MemoryState) AddKitty(tx TxHash, kittyID KittyID, address cipher.Address) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.kitties[kittyID]; ok {
		return fmt.Errorf("kitty of id '%d' already exists",
			kittyID)
	}

	if kState, ok := s.kitties[kittyID]; !ok {
		s.kitties[kittyID] = &KittyState{
			Address:      address,
			Transactions: TxHashes{tx},
		}
	} else {
		kState.Address = address
		kState.Transactions = append(kState.Transactions, tx)
	}

	if aState, ok := s.addresses[address]; !ok {
		s.addresses[address] = &AddressState{
			Kitties:      KittyIDs{kittyID},
			Transactions: TxHashes{tx},
		}
	} else {
		aState.Kitties.Add(kittyID)
		aState.Transactions = append(aState.Transactions, tx)
	}

	return nil
}

func (s *MemoryState) MoveKitty(tx TxHash, kittyID KittyID, from, to cipher.Address) error {
	s.Lock()
	defer s.Unlock()

	if from == to {
		return fmt.Errorf("kitty of id '%d' already belongs to address '%s'",
			kittyID, from)

	} else if kState, ok := s.kitties[kittyID]; !ok {
		return fmt.Errorf("kitty of id '%d' does not exist",
			kittyID)

	} else if kState.Address != from {
		return fmt.Errorf("kitty of id '%d' does not belong to address '%s'",
			kittyID, from)
	}

	kState := s.kitties[kittyID]
	kState.Address = to
	kState.Transactions = append(kState.Transactions, tx)

	if fromState, ok := s.addresses[from]; !ok {
		panic(fmt.Errorf(
			"state of 'from' address '%s' does not exist in state",
			from.String()))
	} else {
		fromState.Kitties.Remove(kittyID)
		fromState.Transactions = append(fromState.Transactions, tx)
	}

	if toState, ok := s.addresses[to]; !ok {
		s.addresses[to] = &AddressState{
			Kitties:      KittyIDs{kittyID},
			Transactions: TxHashes{tx},
		}
	} else {
		toState.Kitties.Add(kittyID)
		toState.Transactions = append(toState.Transactions, tx)
	}
	return nil
}
