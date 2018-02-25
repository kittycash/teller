package box

import "sync"

//@TODO a collections struct?
//@TODO what other info can kitty box store?
//@TODO implement this
// Box represents a unique kitty box
type Box struct {
	KittyID string
}

// Kind represents the type of kitty
// contains the description and can contain other information if required
// @TODO
type Kind struct {
	Description string
}

// BoxType represents a type of box users can buy
type BoxType struct {
	sync.RWMutex
	// Name of the box collection
	name string
	// type of kitty
	kind Kind
	// Price of box in BTC satoshis
	priceBTC uint64
	// Price of box in SKY droplets
	priceSKY uint64
	// Pool of boxes to sell
	boxes   []Box
}