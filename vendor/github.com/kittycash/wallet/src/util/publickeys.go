package util

import "github.com/skycoin/skycoin/src/cipher"

func MustPubKeysFromStrings(pkStrings []string) []cipher.PubKey {
	out := make([]cipher.PubKey, len(pkStrings))
	for i, pkStr := range pkStrings {
		out[i] = cipher.MustPubKeyFromHex(pkStr)
	}
	return out
}
