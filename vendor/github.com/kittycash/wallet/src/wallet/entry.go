package wallet

import (
	"errors"
	"github.com/skycoin/skycoin/src/cipher"
)

// FloatingEntry represents a readable wallet entry.
type FloatingEntry struct {
	Address string `json:"address"`
	PubKey  string `json:"public_key"`
	SecKey  string `json:"secret_key"`
}

// Entry represents a wallet entry.
type Entry struct {
	Address cipher.Address
	PubKey  cipher.PubKey
	SecKey  cipher.SecKey
}

func NewEntry(sk cipher.SecKey) (*Entry, error) {
	if e := sk.Verify(); e != nil {
		return nil, e
	}
	return &Entry{
		Address: cipher.AddressFromSecKey(sk),
		PubKey:  cipher.PubKeyFromSecKey(sk),
		SecKey:  sk,
	}, nil
}

func (we *Entry) ToFloating() *FloatingEntry {
	return &FloatingEntry{
		Address: we.Address.String(),
		PubKey:  we.PubKey.Hex(),
		SecKey:  we.SecKey.Hex(),
	}
}

// Verify checks that the public key is derivable from the secret key,
// and that the public key is associated with the address
func (we *Entry) Verify() error {
	if cipher.PubKeyFromSecKey(we.SecKey) != we.PubKey {
		return errors.New("invalid public key for secret key")
	}
	return we.VerifyPublic()
}

// VerifyPublic checks that the public key is associated with the address
func (we *Entry) VerifyPublic() error {
	if err := we.PubKey.Verify(); err != nil {
		return err
	}
	return we.Address.Verify(we.PubKey)
}
