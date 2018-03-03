package iko

import (
	"errors"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"log"
	"time"
)

type TxHash cipher.SHA256

func (h TxHash) Hex() string {
	return cipher.SHA256(h).Hex()
}

type TxHashes []TxHash

func (h TxHashes) ToStringArray() []string {
	out := make([]string, len(h))
	for i, hash := range h {
		out[i] = hash.Hex()
	}
	return out
}

type TxAction func(tx *Transaction) error

// Transaction represents a kitty transaction.
// For IKO, transaction and block are combined to formed one entity.
type Transaction struct {
	Prev TxHash
	Seq  uint64 // Each transaction has a sequence.
	TS   int64  // Timestamp.

	KittyID KittyID
	From    cipher.Address
	To      cipher.Address
	Sig     cipher.Sig
}

// NewGenTx creates a "gen" transaction. This is where a kitty is created on the blockchain.
func NewGenTx(prev *Transaction, kittyID KittyID, sk cipher.SecKey) *Transaction {
	var (
		address = cipher.AddressFromSecKey(sk)
		ts      = time.Now().UnixNano()
		tx      *Transaction
	)
	if prev == nil {
		tx = &Transaction{
			Prev:    TxHash{},
			Seq:     0,
			TS:      ts,
			KittyID: kittyID,
			From:    address,
			To:      address,
		}
	} else {
		tx = &Transaction{
			Prev:    prev.Hash(),
			Seq:     prev.Seq + 1,
			TS:      ts,
			KittyID: kittyID,
			From:    address,
			To:      address,
		}
	}
	tx.Sig = tx.Sign(sk)
	return tx
}

// NewTransferTx creates a normal transaction where a kitty is transferred from
// one address to another.
func NewTransferTx(prev *Transaction, kittyID KittyID, to cipher.Address, sk cipher.SecKey) *Transaction {
	tx := &Transaction{
		Prev:    prev.Hash(),
		Seq:     prev.Seq + 1,
		TS:      time.Now().UnixNano(),
		KittyID: kittyID,
		From:    cipher.AddressFromSecKey(sk),
		To:      to,
	}
	tx.Sig = tx.Sign(sk)
	return tx
}

func (tx Transaction) Serialize() []byte {
	return encoder.Serialize(tx)
}

func (tx Transaction) Hash() TxHash {
	return TxHash(cipher.SumSHA256(tx.Serialize()))
}

func (tx Transaction) HashInner() cipher.SHA256 {
	tx.Sig = cipher.Sig{}
	return cipher.SHA256(tx.Hash())
}

func (tx Transaction) Sign(sk cipher.SecKey) cipher.Sig {
	e := cipher.
		AddressFromSecKey(sk).
		Verify(cipher.PubKeyFromSecKey(sk))
	if e != nil {
		log.Panic(e)
	}
	return cipher.SignHash(tx.HashInner(), sk)
}

// Verify checks the hash, seq and signature of the transaction.
//		- Previous tx hash.
//		- Tx sequence.
//		- Tx timestamp (needs to be ahead of the previous tx, and behind ts now with threshold).
//		- Tx signature.
// Verify does not check:
//		- Whether from address actually owns the kitty of ID.
//		- Double spending of kitties.
// TODO (evanlinjin): Write tests.
func (tx Transaction) Verify(prev *Transaction) error {
	isGenesis := prev == nil

	// Check hash.
	if isGenesis {
		if tx.Prev != (TxHash{}) {
			return errors.New("invalid prev hash")
		}
	} else {
		if tx.Prev != prev.Hash() {
			return errors.New("invalid prev hash")
		}
	}

	// Check seq.
	if isGenesis {
		if tx.Seq != 0 {
			return errors.New("invalid seq")
		}
	} else {
		if tx.Seq != prev.Seq+1 {
			return errors.New("invalid seq")
		}
	}

	// Check timestamp.
	if prev != nil {
		if tx.TS <= prev.TS || tx.TS > time.Now().UnixNano()+int64(time.Minute) {
			return errors.New("invalid ts")
		}
	}

	// Check signature.
	return cipher.ChkSig(tx.From, tx.HashInner(), tx.Sig)
}

// IsKittyGen returns true if:
//		- Tx is of the correct structure to create a new kitty.
//		- Tx is of the right address to create a new kitty.
// TODO (evanlinjin): Write tests.
func (tx Transaction) IsKittyGen(pk cipher.PubKey) bool {
	// Check from address.
	if e := tx.From.Verify(pk); e != nil {
		return false
	}
	// Check to & from addresses are the same.
	if tx.To != tx.From {
		return false
	}
	// Accept.
	return true
}

// String returns human readable string of transaction.
func (tx Transaction) String() string {
	return fmt.Sprintf("prev:%s|seq:%d|ts:%d|kitty_id:%d|from:%s|to:%s|sig:%s",
		tx.Prev.Hex(), tx.Seq, tx.TS, tx.KittyID, tx.From.String(), tx.To.String(), tx.Sig.Hex())
}
