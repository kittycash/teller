package rand

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kittycash/wallet/src/iko"
	"github.com/skycoin/skycoin/src/cipher"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateKitties(count uint64, sk cipher.SecKey) []*iko.KittyEntry {
	out := make([]*iko.KittyEntry, count)
	for i := range out {
		out[i] = GenerateKitty(iko.KittyID(i), sk)
	}
	return out
}

func GenerateKitty(kittyID iko.KittyID, sk cipher.SecKey) *iko.KittyEntry {

	price := rand.Int63()
	timeSub := rand.Int63n(time.Hour.Nanoseconds() * 24 * 30)

	kitty := iko.Kitty{
		ID:        kittyID,
		Name:      fmt.Sprintf("Kitty No. %d", kittyID),
		Desc:      fmt.Sprintf("Test kitty of ID %d.", kittyID),
		Breed:     GenerateBreed(),
		PriceBTC:  price,
		PriceSKY:  price,
		Created:   time.Now().UnixNano() - timeSub,
		BoxImgURL: fmt.Sprintf("kitty%d.png", kittyID),
	}

	return &iko.KittyEntry{
		Kitty:       kitty,
		Sig:         kitty.Sign(sk).Hex(),
		Reservation: "NONE",
	}
}

func GenerateBreed() string {
	var breeds = []string{
		"rag",
		"police",
		"parrot",
		"shepard",
		"mutated",
		"crazy",
	}
	return breeds[rand.Intn(len(breeds))]
}