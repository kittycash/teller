package wallet

import (
	"fmt"
	"strconv"

	"github.com/skycoin/skycoin/src/cipher/go-bip39"
)

const (
	DefaultSeedBitSize = 128
)

func NewSeed(seedBitSize int) (string, error) {
	entropy, e := bip39.NewEntropy(seedBitSize)
	if e != nil {
		return "", e
	}
	mnemonic, e := bip39.NewMnemonic(entropy)
	if e != nil {
		return "", e
	}
	return mnemonic, nil
}

func ValidSeedBitSizes() []int {
	return []int{DefaultSeedBitSize, 256}
}

// SeedBitSizeFromString determines the requested seed bit size from a
// (possibly POSTed) string. A properly formatted integer string will return
// the integer if it's one of our preconfigured bit sizes. An empty string
// will return the default bit size. Any other strings return an error.
func SeedBitSizeFromString(value string) (int, error) {
	if value == "" {
		return DefaultSeedBitSize, nil
	}
	requestedBitSize, e := strconv.Atoi(value)
	if e != nil {
		return 0, fmt.Errorf("Malformed integer string: %q", value)
	}
	for _, validSize := range ValidSeedBitSizes() {
		if requestedBitSize == validSize {
			return requestedBitSize, nil
		}
	}
	return 0, fmt.Errorf("Unsupported seed bit size: %v (not one of %v)",
		requestedBitSize,
		ValidSeedBitSizes(),
	)
}
