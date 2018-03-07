package box

type KittyID uint64

// Box represents a unique kitty box
type Box struct {
	//KittyID is the unique id for the kitty
	KittyID   string    `json:"kitty_id"`
	// BoxDetail contains metadata for the box
	BoxDetail BoxDetail `json:"box_detail"`
}

// BoxDetail gives more info about the content of the box
type BoxDetail struct {
	// Description defines the kitty inside the box
	Description string `json:"description"`
	// Open defines whether the box was opened or not
	Open bool `json:"open"`
	// Price of box in BTC satoshis
	PriceBTC uint64 `json:"price_btc"`
	// Price of box in SKY droplets
	PriceSKY uint64 `json:"price_sky"`
}
