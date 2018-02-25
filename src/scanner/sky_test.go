package scanner

import (
	"testing"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/require"
	"github.com/kittycash/teller/src/util/testutil"
	"time"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/kittycash/teller/src/util/dbutil"
	"errors"
	"fmt"
)

//@TODO add more tests

var (
	dummySkyBlocksBktName = []byte("blocks")
	errNoSkyBlockHash     = errors.New("no block found for height")
)

type dummySkyrpcclient struct {
	db                           *bolt.DB
	blockHashes                  map[int64]string
	blockCount                   int64
	blockCountError              error
	blockVerboseTxError          error
	blockVerboseTxErrorCallCount int
	blockVerboseTxCallCount      int
}

func openDummySkyDB(t *testing.T) *bolt.DB {
	// Blocks 0 through 180 are stored in this DB
	db, err := bolt.Open("./sky.db", 0600, nil)
	require.NoError(t, err)
	return db
}

func newDummySkyrpcclient(db *bolt.DB) *dummySkyrpcclient {
	return &dummySkyrpcclient{
		db:          db,
		blockHashes: make(map[int64]string),
	}
}

func (dsc *dummySkyrpcclient) GetBlockCount() (int64, error) {
	if dsc.blockCountError != nil {
		// blockCountError is only returned once
		err := dsc.blockCountError
		dsc.blockCountError = nil
		return 0, err
	}

	return dsc.blockCount, nil
}

func (dsc *dummySkyrpcclient) GetBlockVerboseTx(seq uint64) (*visor.ReadableBlock, error) {
	dsc.blockVerboseTxCallCount++
	if dsc.blockVerboseTxCallCount == dsc.blockVerboseTxErrorCallCount {
		return nil, dsc.blockVerboseTxError
	}

	// @TODO: implement this

	//var block *visor.ReadableBlock
	//var tmpBlock *visor.ReadableBlock
	//if err := dsc.db.View(func(tx *bolt.Tx) error {
	//	return dbutil.ForEach(tx, dummySkyBlocksBktName, func(k, v []byte) error {
	//		err := json.Unmarshal(v, tmpBlock)
	//		if err != nil {
	//			return err
	//		}
	//
	//		// find the required block
	//		if tmpBlock.Head.BkSeq == seq {
	//			block = tmpBlock
	//			return nil
	//		}
	//
	//		return errNoSkyBlockHash
	//	})
	//}); err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (dsc *dummySkyrpcclient) Shutdown() {}

func setupSkyScannerWithDB(t *testing.T, skyDB *bolt.DB, db *bolt.DB) *SKYScanner {
	log, _ := testutil.NewLogger(t)

	rpc := newDummySkyrpcclient(skyDB)

	// block 180 is the highest block in the database
	rpc.blockCount = 180

	store, err := NewStore(log, db)
	require.NoError(t, err)

	err = store.AddSupportedCoin(CoinTypeSKY)
	require.NoError(t, err)

	cfg := Config{
		ScanPeriod:            time.Millisecond * 10,
		DepositBufferSize:     2,
		InitialScanHeight:     1,
		ConfirmationsRequired: 0,
	}

	scr, err := NewSKYScanner(log, store, rpc, cfg)
	require.NoError(t, err)

	return scr

}

func setupSkyScannerWithNonExistInitHeight(t *testing.T, skyDB *bolt.DB, db *bolt.DB) *SKYScanner {
	log, _ := testutil.NewLogger(t)

	rpc := newDummySkyrpcclient(skyDB)

	// 180 is the highest block in the test data sky.db
	rpc.blockCount = 180

	store, err := NewStore(log, db)
	require.NoError(t, err)
	err = store.AddSupportedCoin(CoinTypeSKY)
	require.NoError(t, err)

	//@TODO improve this, delete some initial blocks
	// Block -1 doesn't exist in db
	cfg := Config{
		ScanPeriod:            time.Millisecond * 10,
		DepositBufferSize:     5,
		InitialScanHeight:     -1,
		ConfirmationsRequired: 0,
	}
	scr, err := NewSKYScanner(log, store, rpc, cfg)
	require.NoError(t, err)

	return scr
}

func setupSkyScanner(t *testing.T, skyDB *bolt.DB) (*SKYScanner, func()) {
	db, shutdown := testutil.PrepareDB(t)

	scr := setupSkyScannerWithDB(t, skyDB, db)

	return scr, shutdown
}

func testSkyScannerRunProcessedLoop(t *testing.T, scr *SKYScanner, nDeposits int) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		var dvs []DepositNote
		for dv := range scr.GetDeposit() {
			dvs = append(dvs, dv)
			dv.ErrC <- nil
		}

		require.Equal(t, nDeposits, len(dvs))

		// check all deposits
		err := scr.Base.GetStorer().(*Store).db.View(func(tx *bolt.Tx) error {
			for _, dv := range dvs {
				var d Deposit
				err := dbutil.GetBucketObject(tx, DepositBkt, dv.ID(), &d)
				require.NoError(t, err)
				if err != nil {
					return err
				}

				require.True(t, d.Processed)
				require.Equal(t, CoinTypeSKY, d.CoinType)
				require.NotEmpty(t, d.Address)
				if d.Value != 0 {
					require.NotEmpty(t, d.Value)
				}
				require.NotEmpty(t, d.Height)
				require.NotEmpty(t, d.Tx)
			}

			return nil
		})
		require.NoError(t, err)
	}()

	// Wait for at least twice as long as the number of deposits to process
	// If there are few deposits, wait at least 5 seconds
	// This only needs to wait at least 1 second normally, but if testing
	// with -race, it needs to wait 5.
	shutdownWait := scr.Base.(*BaseScanner).Cfg.ScanPeriod * time.Duration(nDeposits*2)
	if shutdownWait < minShutdownWait {
		shutdownWait = minShutdownWait
	}

	time.AfterFunc(shutdownWait, func() {
		scr.Shutdown()
	})

	err := scr.Run()
	require.NoError(t, err)
	<-done
}

func testSkyScannerRun(t *testing.T, scr *SKYScanner) {
	nDeposits := 0


	// This address has:
	// 1 deposit, in block 176
	err := scr.AddScanAddress("v4qF7Ceq276tZpTS3HKsZbDguMAcAGAG1q", CoinTypeSKY)
	require.NoError(t, err)
	nDeposits = nDeposits + 1

	// This address has:
	// 2 deposits in block 117
	err = scr.AddScanAddress("8MQsjc5HYbSjPTZikFZYeHHDtLungBEHYS", CoinTypeSKY)
	require.NoError(t, err)
	nDeposits = nDeposits + 2

	// Make sure that the deposit buffer size is less than the number of deposits,
	// to test what happens when the buffer is full
	fmt.Println(scr.Base.(*BaseScanner).Cfg.DepositBufferSize)
	require.True(t, scr.Base.(*BaseScanner).Cfg.DepositBufferSize < nDeposits)

	testSkyScannerRunProcessedLoop(t, scr, nDeposits)
}

func testSkyScannerRunProcessDeposits(t *testing.T, skyDB *bolt.DB) {
	// Tests that the scanner will scan multiple blocks sequentially, finding
	// all relevant deposits and adding them to the depositC channel.
	// All deposits on the depositC channel will be successfully processed
	// by the channel reader, and the scanner will mark these deposits as
	// "processed".
	scr, shutdown := setupSkyScanner(t, skyDB)
	defer shutdown()

	testSkyScannerRun(t, scr)
}


func testSkyScannerInitialGetBlockHashError(t *testing.T, skyDB *bolt.DB) {
	// Test that scanner.Run() returns an error if the initial GetBlockHash
	// based upon scanner.Base.Cfg.InitialScanHeight fails
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()

	scr := setupSkyScannerWithNonExistInitHeight(t, skyDB, db)

	err := scr.Run()
	require.Error(t, err)
	require.Equal(t, errNoSkyBlockHash, err)
}

func TestSkyScanner(t *testing.T) {
	skyDB := openDummySkyDB(t)
	defer testutil.CheckError(t, skyDB.Close)

	//t.Run("group", func(t *testing.T) {
	//	t.Run("RunProcessDeposits", func(t *testing.T) {
	//		testSkyScannerRunProcessDeposits(t, skyDB)
	//	})
	//
	//
	//})

}


