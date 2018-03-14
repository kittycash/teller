package box

//KittyID is the unique id for the kitty
type KittyID uint64

// Box represents a unique kitty box
type Box struct {
	KittyID string `json:"kitty_id"`
	// Detail contains metadata for the box
	Detail Detail `json:"box_detail"`
}

// Detail gives more info about the content of the box
type Detail struct {
	// Description defines the kitty inside the box
	Description string `json:"description"`
	// Open defines whether the box was opened or not
	Open bool `json:"open"`
	// Price of box in BTC satoshis
	PriceBTC int64 `json:"price_btc"`
	// Price of box in SKY droplets
	PriceSKY int64 `json:"price_sky"`
}
