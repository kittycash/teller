package iko

import (
	"github.com/skycoin/skycoin/src/cipher"
	"gopkg.in/sirupsen/logrus.v1"
	"os"
	"sync"
)

type BlockChainConfig struct {
	CreatorPK cipher.PubKey
	TxAction  TxAction
}

func (cc *BlockChainConfig) Prepare() error {
	if cc.TxAction == nil {
		cc.TxAction = func(tx *Transaction) error {
			return nil
		}
	}
	if e := cc.CreatorPK.Verify(); e != nil {
		return e
	}
	return nil
}

type BlockChain struct {
	c     *BlockChainConfig
	chain ChainDB
	state StateDB
	log   *logrus.Logger
	mux   sync.RWMutex

	wg   sync.WaitGroup
	quit chan struct{}
}

func NewBlockChain(config *BlockChainConfig, chainDB ChainDB, stateDB StateDB) (*BlockChain, error) {
	if e := config.Prepare(); e != nil {
		return nil, e
	}
	bc := &BlockChain{
		c:     config,
		chain: chainDB,
		state: stateDB,
		log: &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		},
		quit: make(chan struct{}),
	}

	if e := bc.InitState(); e != nil {
		return nil, e
	}

	bc.wg.Add(1)
	go bc.service()

	return bc, nil
}

func (bc *BlockChain) InitState() error {

	var prev *Transaction
	for i := uint64(1); i < bc.chain.Len(); i++ {

		// Get transaction.
		tx, e := bc.chain.GetTxOfSeq(i)
		if e != nil {
			return e
		}
		bc.log.WithField("tx", tx.String()).Debugf("InitState (%d)", i)

		// Check hash, seq and sig of tx.
		if e := tx.Verify(prev); e != nil {
			return e
		}

		// If tx is to structured to create a kitty, attempt to add to state.
		// Otherwise, attempt to transfer it's ownership in the state.
		if tx.IsKittyGen(bc.c.CreatorPK) {
			if e := bc.state.AddKitty(tx.Hash(), tx.KittyID, tx.To); e != nil {
				return e
			}
		} else {
			if e := bc.state.MoveKitty(tx.Hash(), tx.KittyID, tx.From, tx.To); e != nil {
				return e
			}
		}
	}
	return nil
}

func (bc *BlockChain) Close() {
	close(bc.quit)
}

func (bc *BlockChain) service() {
	defer bc.wg.Done()

	for {
		select {
		case <-bc.quit:
			return

		case tx := <-bc.chain.TxChan():
			if e := bc.c.TxAction(tx); e != nil {
				panic(e)
			}
		}
	}
}

func (bc *BlockChain) GetHeadTx() (Transaction, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.Head()
}

func (bc *BlockChain) GetTxOfHash(txHash TxHash) (Transaction, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.GetTxOfHash(txHash)
}

func (bc *BlockChain) GetTxOfSeq(seq uint64) (Transaction, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain.GetTxOfSeq(seq)
}

func (bc *BlockChain) GetKittyState(kittyID KittyID) (*KittyState, bool) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.state.GetKittyState(kittyID)
}

func (bc *BlockChain) GetAddressState(address cipher.Address) *AddressState {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.state.GetAddressState(address)
}

func (bc *BlockChain) InjectTx(tx *Transaction) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	var prev *Transaction

	if temp, e := bc.chain.Head(); e == nil {
		prev = &temp
	}

	if e := tx.Verify(prev); e != nil {
		return e
	}

	if tx.IsKittyGen(bc.c.CreatorPK) {
		bc.log.
			WithField("kitty_id", tx.KittyID).
			WithField("address", tx.To.String()).
			Debug("adding kitty to state")

		if e := bc.state.AddKitty(tx.Hash(), tx.KittyID, tx.To); e != nil {
			return e
		}
	} else {
		if e := bc.state.MoveKitty(tx.Hash(), tx.KittyID, tx.From, tx.To); e != nil {
			return e
		}
	}

	if e := bc.chain.AddTx(*tx); e != nil {
		panic(e)
	}

	return nil
}
