package box


type KittyID uint64
//@TODO a collections struct?
//@TODO what other info can kitty box store?
//@TODO implement this

// Box represents a unique kitty box
type Box struct {
	KittyID string `json:"kitty_id"`
	BoxDetail BoxDetail `json:"box_detail"`
	// Open defines whether the box was opened or not
	Open      bool  `json:"open"`
	// Price of box in BTC satoshis
	PriceBTC uint64 `json:"price_btc"`
	// Price of box in SKY droplets
	PriceSKY uint64 `json:"price_sky"`
}


// BoxDetail gives more info about the content of the box
type BoxDetail struct {
	Description string `json:"description"`
}