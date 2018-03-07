package sender

import (
	"errors"

	"fmt"
	"net/http"
	"os"

	kittyClient "github.com/kittycash/wallet/src/http"
	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/skycoin/skycoin/src/cipher"
)

// APIError wraps errors from the kittycash http api
type APIError struct {
	error
}

// NewRPCError wraps an err with RPCError
func NewAPIError(err error) APIError {
	return APIError{err}
}

// API provides methods for sending kitties
type API struct {
	wallet   *wallet.Wallet
	httpAddr string
}

// NewAPI creates an API instance
func NewAPI(wltLabel, wltPassword, httpAddr string) (*API, error) {
	f, err := os.Open(wallet.LabelPath(wltLabel))
	if err != nil {
		return nil, err
	}

	// load a readable wallet
	fw, err := wallet.LoadFloatingWallet(f, wltLabel, wltPassword)
	if err != nil {
		return nil, err
	}

	return &API{
		wallet:   fw,
		httpAddr: httpAddr,
	}, nil
}

// CreateTransaction creates a transfer kitty transaction
func (c *API) CreateTransaction(recvAddr string, kittyID iko.KittyID) (*iko.Transaction, error) {
	// get latest tx
	tx, resp := kittyClient.GetHeadTx(c.httpAddr)
	if resp.Status != http.StatusOK {
		return nil, resp.Error
	}

	toAddr, err := cipher.DecodeBase58Address(recvAddr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to decode %v: %v", recvAddr, err.Error()))
	}

	// get the kitty state which gives us the address that stores kitty
	ks, resp := kittyClient.GetKittyState(c.httpAddr, kittyID)
	if resp.Status != http.StatusOK {
		return nil, resp.Error
	}

	// get the required wallet entry
	var secKey cipher.SecKey
	for _, w := range c.wallet.Entries {
		if w.Address == ks.Address {
			secKey = w.SecKey
			break
		}
	}

	//@TODO (therealssj): we need to move box? don't know how yet.
	// create a transfer tx
	transferTx := iko.NewTransferTx(tx, kittyID, toAddr, secKey)

	return transferTx, nil
}

// InjectTransaction broadcasts a transaction and returns its txhash
func (c *API) InjectTransaction(tx *iko.Transaction) (string, error) {
	resp := kittyClient.InjectTx(c.httpAddr, tx)
	if resp.Status != http.StatusOK {
		return "", resp.Error
	}

	return tx.Hash().Hex(), nil
}

// GetTransaction returns transaction by txhash
func (c *API) GetTransaction(txhash iko.TxHash) (*iko.Transaction, error) {
	txn, resp := kittyClient.GetTxOfHash(c.httpAddr, txhash)
	if resp.Error != nil {
		return nil, APIError{resp.Error}
	}

	return txn, nil
}

// Balance returns the balance of a wallet
func (c *API) Balance() int {
	//@TODO (therealssj): implement this

	// for now we just return the no. of address as balance
	return c.wallet.Count()
}
