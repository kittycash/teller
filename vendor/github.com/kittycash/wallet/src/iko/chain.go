package iko

import (
	"errors"
	"fmt"
	"sync"
)

// ChainDB represents where the transactions/blocks are stored.
// For iko, we combined blocks and transactions to become a single entity.
// Checks for whether txs are malformed shouldn't happen here.
type ChainDB interface {
	// Head should obtain the head transaction.
	// It should return an error when there are no transactions recorded.
	Head() (Transaction, error)

	// HeadSeq should obtain the sequence index of the head transaction.
	HeadSeq() uint64

	// Len should obtain the length of the chain.
	Len() uint64

	// AddTx should add a transaction to the chain.
	AddTx(tx Transaction) error

	// GetTxOfHash should obtain a transaction of a given hash.
	// It should return an error when the tx doesn't exist.
	GetTxOfHash(hash TxHash) (Transaction, error)

	// GetTxOfSeq should obtain a transaction of a given sequence.
	// It should return an error when the sequence given is invalid,
	//	or the tx doesn't exist.
	GetTxOfSeq(seq uint64) (Transaction, error)

	// TxChan obtains a channel where new transactions are sent through.
	// When a transaction is successfully saved to the `ChainDB` implementation,
	//	we expect to see it getting sent through here too.
	TxChan() <-chan *Transaction
}

type MemoryChain struct {
	sync.RWMutex
	txs    []Transaction
	byHash map[TxHash]*Transaction
	txChan chan *Transaction
}

func NewMemoryChain(bufferSize int) *MemoryChain {
	return &MemoryChain{
		byHash: make(map[TxHash]*Transaction),
		txChan: make(chan *Transaction, bufferSize),
	}
}

func (c *MemoryChain) Head() (Transaction, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.txs) == 0 {
		return Transaction{}, errors.New("no transactions")
	}
	return c.txs[len(c.txs)-1], nil
}

func (c *MemoryChain) HeadSeq() uint64 {
	c.RLock()
	defer c.RUnlock()

	return uint64(len(c.txs)) - 1
}

func (c *MemoryChain) Len() uint64 {
	c.RLock()
	defer c.RUnlock()

	return uint64(len(c.txs))
}

func (c *MemoryChain) AddTx(tx Transaction) error {
	c.Lock()
	defer c.Unlock()

	c.txs = append(c.txs, tx)
	c.byHash[tx.Hash()] = &c.txs[len(c.txs)-1]
	go func() {
		c.txChan <- &c.txs[len(c.txs)-1]
	}()
	return nil
}

func (c *MemoryChain) GetTxOfHash(hash TxHash) (Transaction, error) {
	c.Lock()
	defer c.Unlock()

	tx, ok := c.byHash[hash]
	if !ok {
		return Transaction{}, fmt.Errorf("tx of hash '%s' does not exist", hash.Hex())
	}
	return *tx, nil
}

func (c *MemoryChain) GetTxOfSeq(seq uint64) (Transaction, error) {
	c.RLock()
	defer c.RUnlock()

	if seq >= uint64(len(c.txs)) {
		return Transaction{}, fmt.Errorf("block of sequence '%d' does not exist", seq)
	}
	return c.txs[seq], nil
}

func (c *MemoryChain) TxChan() <-chan *Transaction {
	return nil
}
