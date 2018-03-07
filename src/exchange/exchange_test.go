package exchange

import (
	"sync"
	"testing"

	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/kittycash/teller/src/config"
	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/teller/src/util/testutil"
	"github.com/kittycash/wallet/src/iko"
)

type dummySender struct {
	sync.RWMutex
	createTransactionErr    error
	broadcastTransactionErr error
	confirmErr              error
	txidConfirmMap          map[string]bool
	fromAddr                string
}

func newDummySender() *dummySender {
	return &dummySender{
		txidConfirmMap: make(map[string]bool),
		fromAddr:       "nYTKxHm6SZWAMdDVx6U9BqxKMuCjmSLp93",
	}
}

func (s *dummySender) CreateTransaction(destAddr string, kittyID iko.KittyID) (*iko.Transaction, error) {
	if s.createTransactionErr != nil {
		return nil, s.createTransactionErr
	}

	addr := cipher.MustDecodeBase58Address(destAddr)

	return &iko.Transaction{
		KittyID: kittyID,
		To:      addr,
	}, nil
}

func (s *dummySender) BroadcastTransaction(tx *iko.Transaction) *sender.BroadcastTxResponse {
	req := sender.BroadcastTxRequest{
		Tx:   tx,
		RspC: make(chan *sender.BroadcastTxResponse, 1),
	}

	if s.broadcastTransactionErr != nil {
		return &sender.BroadcastTxResponse{
			Err: s.broadcastTransactionErr,
			Req: req,
		}
	}

	return &sender.BroadcastTxResponse{
		Txid: tx.Hash().Hex(),
		Req:  req,
	}
}

func (s *dummySender) IsTxConfirmed(txid *iko.TxHash) *sender.ConfirmResponse {
	s.RLock()
	defer s.RUnlock()

	req := sender.ConfirmRequest{
		TxHash: txid,
	}

	if s.confirmErr != nil {
		return &sender.ConfirmResponse{
			Err: s.confirmErr,
			Req: req,
		}
	}

	confirmed := s.txidConfirmMap[txid.Hex()]
	return &sender.ConfirmResponse{
		Confirmed: confirmed,
		Req:       req,
	}
}

func (s *dummySender) predictTxid(t *testing.T, destAddr string, kittyID iko.KittyID) string {
	tx, err := s.CreateTransaction(destAddr, kittyID)
	require.NoError(t, err)
	return tx.Hash().Hex()
}

func (s *dummySender) setTxConfirmed(txid string) {
	s.Lock()
	defer s.Unlock()

	s.txidConfirmMap[txid] = true
}

func (s *dummySender) Balance() int {
	return 1
}

type dummyScanner struct {
	dvC   chan scanner.DepositNote
	addrs []string
}

func newDummyScanner() *dummyScanner {
	return &dummyScanner{
		dvC: make(chan scanner.DepositNote, 10),
	}
}

func (scan *dummyScanner) AddScanAddress(btcAddr, coinType string) error {
	scan.addrs = append(scan.addrs, btcAddr)
	return nil
}

func (scan *dummyScanner) GetDeposit() <-chan scanner.DepositNote {
	return scan.dvC
}

func (scan *dummyScanner) GetScanAddresses() ([]string, error) {
	return []string{}, nil
}

func (scan *dummyScanner) addDeposit(d scanner.DepositNote) {
	scan.dvC <- d
}

func (scan *dummyScanner) stop() {
	close(scan.dvC)
}

const (
	testMaxDecimals     = 0
	testSkyAddr         = "2Wbi4wvxC4fkTYMsS2f6HaFfW4pafDjXcQW"
	testSkyAddr2        = "hs1pyuNgxDLyLaZsnqzQG9U3DKdJsbzNpn"
	testWalletFile      = "test.wlt"
	dbScanTimeout       = time.Second * 3
	statusCheckTimeout  = time.Second * 3
	statusCheckInterval = time.Millisecond * 10
	statusCheckNilWait  = time.Second
	dbCheckWaitTime     = time.Millisecond * 300
)

var (
	defaultCfg = config.BoxExchanger{
		TxConfirmationCheckWait: time.Millisecond * 100,
		Wallet:                  testWalletFile,
		SendEnabled:             true,
	}
)

func newTestExchange(t *testing.T, log *logrus.Logger, db *bolt.DB) *Exchange {
	store, err := NewStore(log, db)
	require.NoError(t, err)

	bscr := newDummyScanner()
	sscr := newDummyScanner()
	multiplexer := scanner.NewMultiplexer(log)
	err = multiplexer.AddScanner(bscr, scanner.CoinTypeBTC)
	require.NoError(t, err)
	err = multiplexer.AddScanner(sscr, scanner.CoinTypeSKY)
	require.NoError(t, err)

	go testutil.CheckError(t, multiplexer.Multiplex)

	e, err := NewExchange(log, defaultCfg, store, multiplexer, newDummySender())
	require.NoError(t, err)
	return e
}

func setupExchange(t *testing.T, log *logrus.Logger) (*Exchange, func(), func()) {
	db, shutdownDB := testutil.PrepareDB(t)

	e := newTestExchange(t, log, db)

	done := make(chan struct{})
	run := func() {
		err := e.Run()
		require.NoError(t, err)
		close(done)
	}

	shutdown := func() {
		shutdownDB()
		<-done
	}

	return e, run, shutdown
}

func closeMultiplexer(e *Exchange) {
	mp := e.Receiver.(*Receive).multiplexer
	mp.GetScanner(scanner.CoinTypeBTC).(*dummyScanner).stop()
	mp.GetScanner(scanner.CoinTypeSKY).(*dummyScanner).stop()
	mp.Shutdown()
}

func runExchange(t *testing.T) (*Exchange, func(), *logrus_test.Hook) {
	log, hook := testutil.NewLogger(t)
	e, run, shutdown := setupExchange(t, log)
	go run()
	return e, shutdown, hook
}

func runExchangeMockStore(t *testing.T) (*Exchange, func(), *logrus_test.Hook) {
	store := &MockStore{}
	log, hook := testutil.NewLogger(t)

	bscr := newDummyScanner()
	escr := newDummyScanner()
	multiplexer := scanner.NewMultiplexer(log)
	err := multiplexer.AddScanner(bscr, scanner.CoinTypeBTC)
	require.NoError(t, err)
	err = multiplexer.AddScanner(escr, scanner.CoinTypeSKY)
	require.NoError(t, err)

	go testutil.CheckError(t, multiplexer.Multiplex)

	e, err := NewExchange(log, defaultCfg, store, multiplexer, newDummySender())
	require.NoError(t, err)

	done := make(chan struct{})
	run := func() {
		err := e.Run()
		require.NoError(t, err)
		close(done)
	}
	go run()

	shutdown := func() {
		<-done
	}

	return e, shutdown, hook
}

func checkExchangerStatus(t *testing.T, e Exchanger, expectedErr error) {
	// When the expected error is nil, we can only wait a period of time to
	// hope that an error did not appear
	if expectedErr == nil {
		time.Sleep(statusCheckNilWait)
		require.NoError(t, e.Status())
		return
	}

	// When the expected error is not nil, we can wait to see if the error
	// appears, and check that the error matches
	timeoutTicker := time.After(statusCheckTimeout)
loop:
	for {
		select {
		case <-time.Tick(statusCheckInterval):
			err := e.Status()
			if err == nil {
				continue loop
			}
			require.Equal(t, expectedErr, err)
			break loop
		case <-timeoutTicker:
			t.Fatal("Deposit status checking timed out")
			return
		}
	}
}

func TestExchangeRunShutdown(t *testing.T) {
	// Tests a simple start and stop, with no scanner activity
	e, shutdown, _ := runExchange(t)
	defer shutdown()
	defer e.Shutdown()
	closeMultiplexer(e)
}

func TestExchangeRunScannerClosed(t *testing.T) {
	// Tests that there is no problem when the scanner closes
	e, shutdown, _ := runExchange(t)
	defer shutdown()
	defer e.Shutdown()
	closeMultiplexer(e)
}

//@TODO (therealssj): add tests
