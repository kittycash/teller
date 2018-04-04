package rpc

import (
	"errors"

	"github.com/skycoin/skycoin/src/cipher"

	"github.com/kittycash/wallet/src/iko"
)

const (
	PrefixName  = "kc_rpc"
	NetworkName = "tcp"
)

var (
	ErrRemoteQuitDisabled = errors.New("remote quit disabled")
	ErrKittyDoesNotExist  = errors.New("kitty does not exist")
)

type Gateway struct {
	IKO      *iko.BlockChain
	QuitChan chan int
}

func (g *Gateway) Quit(_, _ *struct{}) error {
	if g.QuitChan == nil {
		return ErrRemoteQuitDisabled
	} else {
		return nil
	}
}

type BalancesIn struct {
	Addresses []cipher.Address
}

type BalancesOut struct {
	Count int
	List  iko.KittyIDs
}

func (g *Gateway) Balances(in *BalancesIn, out *BalancesOut) error {
	for _, address := range in.Addresses {
		aState := g.IKO.GetAddressState(address)
		out.Count += len(aState.Kitties)
		out.List = append(out.List, aState.Kitties...)
	}
	return nil
}

type KittyOwnerIn struct {
	KittyID iko.KittyID
}

type KittyOwnerOut struct {
	Address cipher.Address
	Unspent iko.TxHash
}

func (g *Gateway) KittyOwner(in *KittyOwnerIn, out *KittyOwnerOut) error {
	kState, ok := g.IKO.GetKittyState(in.KittyID)
	if !ok {
		return ErrKittyDoesNotExist
	}
	out.Address = kState.Address
	out.Unspent = kState.Transactions[len(kState.Transactions)-1]
	return nil
}

type TransactionIn struct {
	TxHash iko.TxHash
}

type TransactionOut struct {
	Tx        iko.Transaction
	TxMeta    iko.TxMeta
	Confirmed bool
}

func (g *Gateway) Transaction(in *TransactionIn, out *TransactionOut) error {
	txWrap, err := g.IKO.GetTxOfHash(in.TxHash)
	if err != nil {
		return err
	}
	out.Tx = txWrap.Tx
	out.TxMeta = txWrap.Meta
	out.Confirmed = true
	return nil
}

type InjectTxIn struct {
	Tx iko.Transaction
}

type InjectTxOut struct {
	TxHash iko.TxHash
	Meta   iko.TxMeta
}

func (g *Gateway) InjectTx(in *InjectTxIn, out *InjectTxOut) error {
	meta, e := g.IKO.InjectTx(&in.Tx)
	if e != nil {
		return e
	}
	out.TxHash = in.Tx.Hash()
	out.Meta = *meta
	return nil
}
