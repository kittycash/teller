package wallet

import (
	"bytes"
	"errors"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

var (
	ErrInvalidNonce = errors.New("nonce is invalid")
	ErrFileSize     = errors.New("wallet file size is too small")
)

const (
	PrefixSize  = 16
	VersionSize = 8
	NonceSize   = 8
)

type Prefix [16]byte

func NewPrefix(ver uint64, nonce []byte) Prefix {
	if len(nonce) != NonceSize {
		panic(ErrInvalidNonce)
	}
	var out Prefix
	copy(out[:VersionSize], encoder.Serialize(ver))
	copy(out[VersionSize:], nonce)
	return out
}

func ExtractPrefix(raw []byte) (Prefix, []byte, error) {
	if len(raw) < PrefixSize {
		return Prefix{}, nil, ErrFileSize
	}
	var prefix Prefix
	copy(prefix[:], raw[:PrefixSize])
	return prefix, raw[PrefixSize:], nil
}

func (p Prefix) Version() uint64 {
	var ver uint64
	encoder.DeserializeRaw(p[:VersionSize], &ver)
	return ver
}

func (p Prefix) Nonce() []byte {
	return p[VersionSize:]
}

func (p Prefix) Encrypted() bool {
	return bytes.Equal(p.Nonce(), EmptyNonce()) == false
}

/*
	<<< Nonce >>>
*/

func EmptyNonce() []byte {
	return make([]byte, NonceSize)
}

func RandNonce() []byte {
	var (
		out   = cipher.RandByte(NonceSize)
		empty = make([]byte, NonceSize)
	)
	for bytes.Equal(out, empty) {
		out = cipher.RandByte(NonceSize)
	}
	return out
}
