package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kittycash/wallet/src/iko"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"io/ioutil"
	"net/http"
	"path"
)

type RespMeta struct {
	GotResp bool  // whether we got a reply from the server
	Status  int   // http status code
	Error   error // error (if any)
}

func GetKittyState(httpAddr string, kittyID iko.KittyID) (*iko.KittyState, *RespMeta) {
	r, e := http.DefaultClient.Get(
		path.Join(httpAddr, "/api/iko/kitty/", fmt.Sprintf("%d.enc", kittyID)),
	)
	if e != nil {
		return nil, &RespMeta{
			false, -1, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		kState := new(iko.KittyState)
		if e := encoder.DeserializeRaw(raw, kState); e != nil {
			return nil, &RespMeta{
				true, r.StatusCode, e,
			}
		}
		return kState, &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return nil, &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}

func GetAddressState(httpAddr string, address cipher.Address) (*iko.AddressState, *RespMeta) {
	r, e := http.DefaultClient.Get(
		path.Join(httpAddr, "/api/iko/address/", fmt.Sprintf("%s.enc", address.String())),
	)
	if e != nil {
		return nil, &RespMeta{
			false, -1, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		aState := new(iko.AddressState)
		if e := encoder.DeserializeRaw(raw, aState); e != nil {
			return nil, &RespMeta{
				true, r.StatusCode, e,
			}
		}
		return aState, &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return nil, &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}

func GetTxOfHash(httpAddr string, txHash iko.TxHash) (*iko.Transaction, *RespMeta) {
	r, e := http.DefaultClient.Get(
		path.Join(httpAddr, "/api/iko/tx/", fmt.Sprintf("%s.enc?request=hash", txHash.Hex())),
	)
	if e != nil {
		return nil, &RespMeta{
			false, -1, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		tx := new(iko.Transaction)
		if e := encoder.DeserializeRaw(raw, tx); e != nil {
			return nil, &RespMeta{
				true, r.StatusCode, e,
			}
		}
		return tx, &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return nil, &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}

func GetTxOfSeq(httpAddr string, txSeq uint64) (*iko.Transaction, *RespMeta) {
	r, e := http.DefaultClient.Get(
		path.Join(httpAddr, "/api/iko/tx/", fmt.Sprintf("%d.enc?request=seq", txSeq)),
	)
	if e != nil {
		return nil, &RespMeta{
			false, -1, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		tx := new(iko.Transaction)
		if e := encoder.DeserializeRaw(raw, tx); e != nil {
			return nil, &RespMeta{
				true, r.StatusCode, e,
			}
		}
		return tx, &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return nil, &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}

// GetHeadTx obtains the head transaction.
func GetHeadTx(httpAddr string) (*iko.Transaction, *RespMeta) {
	r, e := http.DefaultClient.Get(
		path.Join(httpAddr, "/api/iko/head_tx.enc"),
	)
	if e != nil {
		return nil, &RespMeta{
			false, r.StatusCode, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		tx := new(iko.Transaction)
		if e := encoder.DeserializeRaw(raw, tx); e != nil {
			return nil, &RespMeta{
				true, r.StatusCode, e,
			}
		}
		return tx, &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return nil, &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}

// InjectTx injects a transaction to the blockchain.
// Note that the transaction needs to be signed and the fields need to be correct. (eg. seq, prev)
func InjectTx(httpAddr string, tx *iko.Transaction) *RespMeta {
	r, e := http.DefaultClient.Post(
		path.Join(httpAddr, "/api/iko/inject_tx"),
		"application/octet-stream",
		bytes.NewReader(tx.Serialize()),
	)
	if e != nil {
		return &RespMeta{
			false, -1, e,
		}
	}
	raw, _ := ioutil.ReadAll(r.Body)
	switch r.StatusCode {
	case http.StatusOK:
		return &RespMeta{
			true, r.StatusCode, nil,
		}
	default:
		return &RespMeta{
			true, r.StatusCode, errors.New(string(raw)),
		}
	}
}
