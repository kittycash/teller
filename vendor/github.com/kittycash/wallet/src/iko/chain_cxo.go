package iko

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/skycoin/cxo/node"
	"github.com/skycoin/cxo/skyobject"
	"github.com/skycoin/cxo/skyobject/registry"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"gopkg.in/sirupsen/logrus.v1"

	"github.com/kittycash/wallet/src/util"
)

type MeasureChain func() uint64

type CXOStore struct {
	Meta  []byte
	Txs   registry.Refs `skyobject:"schema=iko.Transaction"`
	Metas registry.Refs `skyobject:"schema=iko.TxMeta"`
}

var (
	cxoReg = registry.NewRegistry(func(r *registry.Reg) {
		r.Register("cipher.Address", cipher.Address{})
		r.Register("iko.Transaction", Transaction{})
		r.Register("iko.TxMeta", TxMeta{})
		r.Register("iko.Store", CXOStore{})
	})
)

type CXOChainConfig struct {
	Dir                string
	Public             bool
	Memory             bool
	MessengerAddresses []string
	CXOAddress         string
	CXORPCAddress      string

	MasterRooter    bool
	MasterRootPK    cipher.PubKey
	MasterRootSK    cipher.SecKey
	MasterRootNonce uint64 // Public
}

func (c *CXOChainConfig) Process(log *logrus.Logger) error {
	if e := c.MasterRootPK.Verify(); e != nil {
		return e
	}

	if c.MasterRooter {
		if e := c.MasterRootSK.Verify(); e != nil {
			return e
		}
		if c.MasterRootPK != cipher.PubKeyFromSecKey(c.MasterRootSK) {
			return errors.New("public and secret keys do not match")
		}
	}

	if c.Memory {
		c.Dir = ""
	}

	return nil
}

type CXOChain struct {
	mux      sync.Mutex
	c        *CXOChainConfig
	l        *logrus.Logger
	node     *node.Node
	wg       sync.WaitGroup
	received chan *TxWrapper
	accepted chan *TxWrapper

	len util.SafeInt
}

func NewCXOChain(config *CXOChainConfig, modifyNC ...NodeConfigModifier) (*CXOChain, error) {
	log := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	if e := config.Process(log); e != nil {
		return nil, e
	}

	chain := &CXOChain{
		c:        config,
		l:        log,
		received: make(chan *TxWrapper),
		accepted: make(chan *TxWrapper),
	}

	var modify NodeConfigModifier
	if len(modifyNC) > 0 {
		modify = modifyNC[0]
	}

	if e := prepareNode(chain, modify); e != nil {
		return nil, e
	}

	if e := initChain(chain); e != nil {
		log.WithError(e).
			Info("no blockchain yet")
	} else {
		log.WithField("height", chain.len.Val()).
			Info("blockchain re-initialized")
	}

	return chain, nil
}

func (c *CXOChain) Close() {
	defer c.lock()()

	close(c.received)
	c.wg.Wait()
	close(c.accepted)

	if e := c.node.Close(); e != nil {
		c.l.WithError(e).
			Error("error on cxo node close")
	}
	c.l.Print("closed CXO chain DB")
}

/*
	<<< PREP AND SERVICE FUNCTIONS >>>
*/

func (c *CXOChain) lock() func() {
	c.mux.Lock()
	return c.mux.Unlock
}

func (c *CXOChain) attemptPushAccepted(txWrap *TxWrapper) {
	select {
	case c.accepted <- txWrap:
	default:
	}
}

type NodeConfigModifier func(nc *node.Config) error

func prepareNode(chain *CXOChain, modifier NodeConfigModifier) error {

	nc := node.NewConfig()

	nc.DataDir = chain.c.Dir
	nc.Public = chain.c.Public
	nc.InMemoryDB = chain.c.Memory

	nc.TCP.Listen = chain.c.CXOAddress
	if len(chain.c.MessengerAddresses) > 0 {
		nc.TCP.Discovery = node.Addresses(chain.c.MessengerAddresses)
	}
	nc.RPC = chain.c.CXORPCAddress

	nc.OnRootReceived = func(c *node.Conn, r *registry.Root) error {
		defer chain.lock()()

		switch {
		case r.Pub != chain.c.MasterRootPK:
			e := errors.New("received root is not of master public key")
			chain.l.
				WithField("master_pk", chain.c.MasterRootPK.Hex()).
				WithField("received_pk", r.Pub.Hex()).
				Warning(e.Error())
			return e

		case r.Nonce != chain.c.MasterRootNonce:
			e := errors.New("received root is not of master nonce")
			chain.l.
				WithField("master_nonce", chain.c.MasterRootNonce).
				WithField("received_nonce", r.Nonce).
				Warning(e.Error())
			return e

		case len(r.Refs) <= 0:
			e := errors.New("empty refs")
			chain.l.Warning(e.Error())
			return e

		default:
			chain.l.Info("blockchain syncing")
			return nil
		}
	}

	nc.OnRootFilled = func(n *node.Node, r *registry.Root) {
		defer chain.lock()()

		e := func(c *CXOChain, n *node.Node, r *registry.Root) error {
			var store = new(CXOStore)

			p, e := n.Container().Pack(r, cxoReg)
			if e != nil {
				return e
			}

			if e := r.Refs[0].Value(p, store); e != nil {
				return e
			}

			rLen, e := store.Txs.Len(p)
			if e != nil {
				return e
			}

			switch {
			case rLen < c.len.Val():
				return errors.New("received new root has less transactions")

			case rLen == c.len.Val():
				c.l.Info("received new root has no new transactions")
				return nil
			}

			for i := c.len.Val(); i < rLen; i++ {

				var wrapper = new(TxWrapper)

				txHash, e := store.Txs.ValueByIndex(p, int(i), &wrapper.Tx)
				if e != nil {
					return e
				}

				_, e = store.Metas.ValueByIndex(p, int(i), &wrapper.Meta)
				if e != nil {
					return e
				}

				c.l.
					WithField("tx_hash", txHash.Hex()).
					WithField("tx_seq", i).
					Info("received new transaction")

				c.received <- wrapper
			}

			chain.l.Info("blockchain synced")
			return nil

		}(chain, n, r)

		if e != nil {
			chain.l.Error(e.Error())
			return
		}
	}

	nc.OnConnect = func(c *node.Conn) error {
		defer chain.lock()()

		// TODO: implement.
		return nil
	}

	nc.OnDisconnect = func(c *node.Conn, reason error) {
		defer chain.lock()()

		// TODO: implement.
	}

	if modifier != nil {
		if e := modifier(nc); e != nil {
			return e
		}
	}

	var e error
	if chain.node, e = node.NewNode(nc); e != nil {
		return e
	}

	if e := chain.node.Share(chain.c.MasterRootPK); e != nil {
		return e
	}

	return nil
}

func (c *CXOChain) RunTxService(txChecker TxChecker) error {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		for {
			select {
			case txWrapper, ok := <-c.received:
				if !ok {
					return
				} else if err := txChecker(&txWrapper.Tx); err != nil {
					c.l.Println(err.Error())
				} else {
					c.len.Inc()
					c.attemptPushAccepted(txWrapper)
				}
			}
		}
	}()
	return nil
}

func initChain(c *CXOChain) error {
	defer c.lock()()

	r, e := cxoRoot(c)
	if e != nil {
		return e
	}

	p, e := cxoPack(c, r)
	if e != nil {
		return e
	}

	store, e := cxoGetStore(r, p)
	if e != nil {
		return e
	}

	sLen, e := store.Txs.Len(p)
	if e != nil {
		return e
	}

	c.len.Set(sLen)
	return nil
}

/*
	<<< PUBLIC FUNCTIONS >>>
*/

func (c *CXOChain) MasterInitChain() error {
	defer c.lock()()

	up, e := cxoUnpack(c)
	if e != nil {
		return e
	}

	s := new(CXOStore)
	sReg, e := cxoNewStore(up, s)
	if e != nil {
		return e
	}

	r := &registry.Root{
		Refs:  []registry.Dynamic{sReg},
		Reg:   cxoReg.Reference(),
		Pub:   c.c.MasterRootPK,
		Nonce: c.c.MasterRootNonce,
	}

	if e := c.node.Container().Save(up, r); e != nil {
		return e
	}

	c.node.Publish(r)

	return nil
}

func (c *CXOChain) Head() (TxWrapper, error) {
	defer c.lock()()

	var (
		txWrap TxWrapper
		cLen   = c.len.Val()
	)

	if cLen < 1 {
		return txWrap, errors.New("no transactions available")
	}

	store, _, p, e := c.getStore(gsRead)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Txs.ValueByIndex(p, cLen-1, &txWrap.Tx); e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, cLen-1, &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXOChain) Len() uint64 {
	defer c.lock()()
	return uint64(c.len.Val())
}

func (c *CXOChain) AddTx(txWrap TxWrapper, check TxChecker) error {

	if c.c.MasterRooter == false {
		return errors.New("not master node")
	}
	if e := check(&txWrap.Tx); e != nil {
		c.l.WithError(e).Error("failed")
		return e
	}

	defer c.lock()()
	cLen := c.len.Val()

	store, r, up, e := c.getStore(gsWrite)
	if e != nil {
		return e
	}
	if e := store.Txs.AppendValues(up, txWrap.Tx); e != nil {
		return e
	}
	if e := store.Metas.AppendValues(up, txWrap.Meta); e != nil {
		return e
	}
	if e := r.Refs[0].SetValue(up, store); e != nil {
		return e
	}
	if e := c.node.Container().Save(up.(*skyobject.Unpack), r); e != nil {
		return e
	}
	c.node.Publish(r)
	c.len.Set(cLen + 1)
	c.attemptPushAccepted(&txWrap)
	return nil
}

func (c *CXOChain) GetTxOfHash(hash TxHash) (TxWrapper, error) {
	defer c.lock()()
	var txWrap TxWrapper

	store, _, p, e := c.getStore(gsRead)
	if e != nil {
		return txWrap, e
	}
	i, e := store.Txs.ValueOfHashWithIndex(p, cipher.SHA256(hash), &txWrap.Tx)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, i, &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXOChain) GetTxOfSeq(seq uint64) (TxWrapper, error) {
	defer c.lock()()
	var txWrap TxWrapper

	store, _, p, e := c.getStore(gsRead)
	if e != nil {
		return txWrap, e
	}
	if _, e := store.Txs.ValueByIndex(p, int(seq), &txWrap.Tx); e != nil {
		return txWrap, e
	}
	if _, e := store.Metas.ValueByIndex(p, int(seq), &txWrap.Meta); e != nil {
		return txWrap, e
	}
	return txWrap, nil
}

func (c *CXOChain) TxChan() <-chan *TxWrapper {
	return c.accepted
}

func (c *CXOChain) GetTxsOfSeqRange(startSeq uint64, pageSize uint64) ([]TxWrapper, error) {
	defer c.lock()()
	var txWraps []TxWrapper

	if pageSize == 0 {
		return txWraps, fmt.Errorf("invalid pageSize: %d", pageSize)
	}
	cLen := uint64(c.len.Val())
	if startSeq >= cLen {
		return txWraps, fmt.Errorf("invalid startSeq: %d", startSeq)
	}
	if startSeq+pageSize > cLen {
		diff := startSeq + pageSize - cLen
		if pageSize-diff <= 0 {
			return []TxWrapper{}, nil
		}
		pageSize -= diff
	}
	store, _, p, e := c.getStore(gsRead)
	if e != nil {
		return txWraps, e
	}
	txRefs, e := store.Txs.Slice(p, int(startSeq), int(startSeq+pageSize))
	if e != nil {
		return txWraps, e
	}
	refsLen, e := txRefs.Len(p)
	if e != nil {
		return txWraps, e
	}
	metaRefs, e := store.Metas.Slice(p, int(startSeq), int(startSeq+pageSize))
	if e != nil {
		return txWraps, e
	}
	txWraps = make([]TxWrapper, refsLen)
	e = txRefs.Ascend(p, func(i int, hash cipher.SHA256) error {
		raw, _, e := c.node.Container().Get(hash, 0)
		if e != nil {
			return e
		}
		return encoder.DeserializeRaw(raw, &txWraps[i].Tx)
	})
	if e != nil {
		return txWraps, e
	}
	e = metaRefs.Ascend(p, func(i int, hash cipher.SHA256) error {
		raw, _, e := c.node.Container().Get(hash, 0)
		if e != nil {
			return e
		}
		return encoder.DeserializeRaw(raw, &txWraps[i].Meta)
	})
	return txWraps, e
}

type getStoreType int

const (
	gsRead  getStoreType = iota
	gsWrite getStoreType = iota
)

func (c *CXOChain) getStore(t getStoreType) (*CXOStore, *registry.Root, registry.Pack, error) {
	r, e := cxoRoot(c)
	if e != nil {
		return nil, nil, nil, e
	}
	var p registry.Pack
	switch t {
	case gsRead:
		if p, e = cxoPack(c, r); e != nil {
			return nil, nil, nil, e
		}
	case gsWrite:
		if p, e = cxoUnpack(c); e != nil {
			return nil, nil, nil, e
		}
	default:
		panic("invalid getStoreType")
	}
	store, e := cxoGetStore(r, p)
	if e != nil {
		return nil, nil, nil, e
	}
	return store, r, p, nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func cxoRoot(c *CXOChain) (*registry.Root, error) {
	r, e := c.node.Container().LastRoot(c.c.MasterRootPK, c.c.MasterRootNonce)
	if e != nil {
		return nil, e
	}
	return r, nil
}

func cxoGetStore(r *registry.Root, p registry.Pack) (*CXOStore, error) {
	if len(r.Refs) < 1 {
		return nil, errors.New("corrupt root, invalid ref count")
	}
	store := new(CXOStore)
	if e := r.Refs[0].Value(p, store); e != nil {
		return nil, e
	}
	return store, nil
}

func cxoPack(c *CXOChain, r *registry.Root) (*skyobject.Pack, error) {
	p, e := c.node.Container().Pack(r, cxoReg)
	if e != nil {
		return nil, e
	}

	return p, nil
}

func cxoUnpack(c *CXOChain) (*skyobject.Unpack, error) {
	if c.c.MasterRooter == false {
		return nil, errors.New("not master")
	}
	return c.node.Container().Unpack(c.c.MasterRootSK, cxoReg)
}

func cxoNewStore(up *skyobject.Unpack, s *CXOStore) (registry.Dynamic, error) {
	raw := encoder.Serialize(s)
	hash, e := up.Add(raw)
	if e != nil {
		return registry.Dynamic{}, e
	}
	schema, e := up.Registry().SchemaByName("iko.Store")
	if e != nil {
		return registry.Dynamic{}, e
	}
	return registry.Dynamic{
		Hash:   hash,
		Schema: schema.Reference(),
	}, nil
}
