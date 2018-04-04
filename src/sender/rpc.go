package sender

import (
	"fmt"

	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/rpc"

	"github.com/skycoin/skycoin/src/cipher"
)

// RPCError wraps errors from the kittycash RPC library
type RPCError struct {
	error
}

// NewRPCError wraps an err with RPCError
func NewRPCError(err error) RPCError {
	return RPCError{err}
}

// RPC provides methods for sending kitties
type RPC struct {
	rpcAddr   string
	rpcClient *rpc.Client
	kittyAddr string
}

// NewRPC creates an RPC instance
func NewRPC(rpcAddr string) (*RPC, error) {
	client, err := rpc.NewClient(&rpc.ClientConfig{
		Address: rpcAddr,
	})
	if err != nil {
		return nil, err
	}
	return &RPC{
		rpcAddr:   rpcAddr,
		rpcClient: client,
	}, nil
}

// CreateTransaction creates a transfer kitty transaction
func (c *RPC) CreateTransaction(recvAddr string, kittyID iko.KittyID, key cipher.SecKey) (*iko.Transaction, error) {
	kittyOwner, err := c.rpcClient.KittyOwner(&rpc.KittyOwnerIn{
		KittyID: kittyID,
	})
	if err != nil {
		return nil, err
	}

	toAddr, err := cipher.DecodeBase58Address(recvAddr)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode %v: %v", recvAddr, err.Error())
	}

	// create a transaction and sign it using the genesis secret key
	inTx := iko.Transaction{
		KittyID: kittyID,
		In:      kittyOwner.Unspent,
		Out:     kittyOwner.Address,
	}
	inTx.Sig = inTx.Sign(key)

	// create a transfer tx
	transferTx, err := iko.NewTransferTx(&inTx, toAddr, key)
	if err != nil {
		return nil, err
	}
	return transferTx, nil
}

// GetTransaction returns transaction by txhash
func (c *RPC) GetTransaction(txHash iko.TxHash) (*iko.Transaction, error) {
	txnIn := &rpc.TransactionIn{
		TxHash: txHash,
	}
	txn, err := c.rpcClient.Transaction(txnIn)
	if err != nil {
		return nil, err
	}

	return &txn.Tx, nil
}

// InjectTransaction broadcasts a transaction and returns its seq
func (c *RPC) InjectTransaction(tx *iko.Transaction) (string, error) {
	txOut, err := c.rpcClient.InjectTx(&rpc.InjectTxIn{
		Tx: *tx,
	})
	if err != nil {
		return "", err
	}

	return txOut.TxHash.Hex(), nil

}

// Balance returns the balance of an address
func (c *RPC) Balance() (int, error) {
	addr, err := cipher.DecodeBase58Address(c.kittyAddr)
	if err != nil {
		return 0, err
	}

	balance, err := c.rpcClient.Balances(&rpc.BalancesIn{
		Addresses: []cipher.Address{addr},
	})
	if err != nil {
		return 0, err
	}

	return balance.Count, nil
}
