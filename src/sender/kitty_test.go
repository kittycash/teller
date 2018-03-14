package sender

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kittycash/wallet/src/iko"

	"github.com/skycoin/skycoin/src/cipher"

	"github.com/skycoin/teller/src/util/testutil"
)

type dummyKittyClient struct {
	sync.Mutex
	broadcastTxTxid string
	broadcastTxErr  error
	createTxErr     error
	txConfirmed     bool
	getTxErr        error
}

func newDummyKittyClient() *dummyKittyClient {
	return &dummyKittyClient{}
}

func (ds *dummyKittyClient) InjectTransaction(tx *iko.Transaction) (string, error) {
	ds.Lock()
	defer ds.Unlock()
	return ds.broadcastTxTxid, ds.broadcastTxErr
}

func (ds *dummyKittyClient) CreateTransaction(destAddr string, kittyID iko.KittyID, key cipher.SecKey) (*iko.Transaction, error) {
	ds.Lock()
	defer ds.Unlock()
	if ds.createTxErr != nil {
		return nil, ds.createTxErr
	}

	return ds.createTransaction(destAddr, kittyID, key)
}

func (ds *dummyKittyClient) createTransaction(destAddr string, kittyID iko.KittyID, key cipher.SecKey) (*iko.Transaction, error) {
	addr, err := cipher.DecodeBase58Address(destAddr)
	if err != nil {
		return nil, err
	}

	txn := iko.Transaction{
		Out:     addr,
		KittyID: kittyID,
	}
	txn.Sig = txn.Sign(key)

	return &txn, nil
}

func (ds *dummyKittyClient) GetTransaction(txhash iko.TxHash) (*iko.Transaction, error) {
	ds.Lock()
	defer ds.Unlock()
	txJSON := iko.Transaction{}

	return &txJSON, ds.getTxErr
}

func (ds *dummyKittyClient) Balance() (int, error) {
	return 1, nil
}

func (ds *dummyKittyClient) changeConfirmStatus(v bool) {
	ds.Lock()
	defer ds.Unlock()
	ds.txConfirmed = v
}

func (ds *dummyKittyClient) changeBroadcastTxErr(err error) {
	ds.Lock()
	defer ds.Unlock()
	ds.broadcastTxErr = err
}

func (ds *dummyKittyClient) changeBroadcastTxTxid(txid string) { // nolint: unparam
	ds.Lock()
	defer ds.Unlock()
	ds.broadcastTxTxid = txid
}

func (ds *dummyKittyClient) changeGetTxErr(err error) {
	ds.Lock()
	defer ds.Unlock()
	ds.getTxErr = err
}

func TestSenderBroadcastTransaction(t *testing.T) {
	log, _ := testutil.NewLogger(t)
	dsc := newDummyKittyClient()

	dsc.changeBroadcastTxTxid("1111")
	secKey := cipher.SecKey([32]byte{
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
		3, 4, 5, 6,
	})

	s := NewService(log, dsc, secKey)
	go func() {
		err := s.Run()
		require.NoError(t, err)
	}()

	addr := "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7"
	sdr := NewRetrySender(s)

	broadcastTx := func(sender Sender, addr string, amt uint64) (string, error) {
		tx, err := sdr.CreateTransaction(addr, 2)
		if err != nil {
			return "", err
		}

		rsp := sdr.BroadcastTransaction(tx)
		require.NotNil(t, rsp)

		if rsp.Err != nil {
			return "", rsp.Err
		}

		return rsp.Txid, nil
	}

	t.Log("=== Run\tTest broadcastTx normal")
	time.AfterFunc(500*time.Millisecond, func() {
		dsc.changeConfirmStatus(true)
	})
	txid, err := broadcastTx(sdr, addr, 10)
	require.Nil(t, err)
	require.Equal(t, "1111", txid)

	// test broadcastTx coin failed
	t.Log("=== Run\tTest broadcastTx failed")
	dsc.changeConfirmStatus(false)
	dsc.changeBroadcastTxErr(errors.New("connect to node failed"))
	time.AfterFunc(5*time.Second, func() {
		dsc.changeBroadcastTxErr(nil)
		dsc.changeConfirmStatus(true)
	})

	txid, err = broadcastTx(sdr, addr, 20)
	require.Nil(t, err)
	require.Equal(t, "1111", txid)

	// test get transaction failed
	t.Log("=== Run\ttest transaction falied")
	dsc.changeConfirmStatus(false)
	dsc.getTxErr = errors.New("get transaction failed")
	time.AfterFunc(5*time.Second, func() {
		dsc.changeGetTxErr(nil)
	})

	time.AfterFunc(7*time.Second, func() {
		dsc.changeConfirmStatus(true)
	})

	txid, err = broadcastTx(sdr, addr, 20)
	require.Nil(t, err)
	require.Equal(t, "1111", txid)

	t.Log("=== Run\tTest invalid request address")
	txid, err = broadcastTx(sdr, "babcd", 20)
	require.Equal(t, "Invalid address length", err.Error())
	require.Empty(t, txid)

	t.Log("=== Run\tTest invalid request address 2")
	txid, err = broadcastTx(sdr, " bxpUG8sCjeT6X1ES5SbD2LZrRudqiTY7wx", 20)
	fmt.Println(err)
	require.Equal(t, "Invalid base58 character", err.Error())
	require.Empty(t, txid)

	t.Log("=== Run\tTest invalid request address 3")
	txid, err = broadcastTx(sdr, "bxpUG8sCjeT6X1ES5SbD2LZrRudqiTY7wxx", 20)
	require.Equal(t, "Invalid address length", err.Error())
	require.Empty(t, txid)
}
