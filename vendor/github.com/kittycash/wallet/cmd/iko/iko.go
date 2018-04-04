package main

import (
	"io/ioutil"
	"os"

	"github.com/skycoin/skycoin/src/cipher"
	"gopkg.in/sirupsen/logrus.v1"
	"gopkg.in/urfave/cli.v1"

	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/rpc"
	"github.com/kittycash/wallet/src/util"
)

const (
	DefaultCXOAddress = "127.0.0.1:7900"
	DefaultRPCAddress = "127.0.0.1:7907"
)

const (
	fInit      = "init"
	fRootPK    = "root-pk"
	fRootSK    = "root-sk"
	fRootNc    = "root-nc"
	fGenTxPK   = "tx-gen-pk"
	fTranTxPKs = "tx-tran-pks"

	fTestMode       = "test"
	fTestTxGenCount = "test-tx-gen-count"
	fTestTxGenSK    = "test-tx-gen-sk"

	fCXODir             = "cxo-dir"
	fCXOAddress         = "cxo-address"
	fCXORPCAddress      = "cxo-rpc-address"
	fDiscoveryAddresses = "messenger-addresses"

	fRPCAddress  = "rpc-address"
	fRemoteClose = "remote-close"
)

func Flag(flag string, short ...string) string {
	if len(short) == 0 {
		return flag
	}
	return flag + ", " + short[0]
}

var (
	app = cli.NewApp()
	log = logrus.New()
)

func init() {
	app.Name = "iko"
	app.Description = "kittycash initial coin offering service"
	app.Flags = cli.FlagsByName{
		/*
			<<< MASTER >>>
		*/
		cli.StringFlag{
			Name:  Flag(fRootPK, "rpk"),
			Usage: "public key to use as main blockchain signer",
		},
		cli.StringFlag{
			Name:  Flag(fRootSK, "rsk"),
			Usage: "secret key to use as main blockchain signer",
		},
		cli.Uint64Flag{
			Name:  Flag(fRootNc, "rn"),
			Usage: "nonce to use as main blockchain identifier",
		},
		cli.StringFlag{
			Name:  Flag(fGenTxPK, "gpk"),
			Usage: "public key that is trusted for generation transactions",
		},
		cli.StringSliceFlag{
			Name:  Flag(fTranTxPKs, "tpk"),
			Usage: "public keys that are trusted for transfer transactions",
		},
		cli.BoolFlag{
			Name:  Flag(fInit),
			Usage: "whether to init the root if it doesn't exist",
		},
		/*
			<<< TEST MODE >>>
		*/
		cli.BoolFlag{
			Name:  Flag(fTestMode, "t"),
			Usage: "whether to use test data for run",
		},
		cli.IntFlag{
			Name:  Flag(fTestTxGenCount, "tc"),
			Usage: "only valid in test mode, injects a number of initial transactions for testing",
		},
		cli.StringFlag{
			Name:  Flag(fTestTxGenSK, "tsk"),
			Usage: "secret key for signing test transactions",
			Value: new(cipher.SecKey).Hex(),
		},
		/*
			<<< CXO CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(fCXODir),
			Usage: "directory to store cxo files",
			Value: "./kc/cxo",
		},
		cli.StringFlag{
			Name:  Flag(fCXOAddress),
			Usage: "address to use to serve CXO",
			Value: DefaultCXOAddress,
		},
		cli.StringSliceFlag{
			Name:  Flag(fDiscoveryAddresses),
			Usage: "discovery addresses",
		},
		cli.StringFlag{
			Name:  Flag(fCXORPCAddress),
			Usage: "address for CXO RPC, leave blank to disable CXO RPC",
		},
		/*
			<<< RPC SERVER >>>
		*/
		cli.StringFlag{
			Name:  Flag(fRPCAddress),
			Usage: "address used to serve rpc, keep empty to not serve rpc",
			Value: DefaultRPCAddress,
		},
		cli.BoolFlag{
			Name:  Flag(fRemoteClose),
			Usage: "whether to enable remote close",
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := util.CatchInterrupt()

	var (
		rootPK     = cipher.MustPubKeyFromHex(ctx.String(fRootPK))
		rootSK     = cipher.MustSecKeyFromHex(ctx.String(fRootSK))
		rootNonce  = ctx.Uint64(fRootNc)
		txGenPK    = cipher.MustPubKeyFromHex(ctx.String(fGenTxPK))
		txTransPKs = util.MustPubKeysFromStrings(ctx.StringSlice(fTranTxPKs))
		doInit     = ctx.Bool(fInit)

		testMode  = ctx.Bool(fTestMode)
		testCount = ctx.Int(fTestTxGenCount)
		testSK    = cipher.MustSecKeyFromHex(ctx.String(fTestTxGenSK))

		cxoDir             = ctx.String(fCXODir)
		cxoAddress         = ctx.String(fCXOAddress)
		cxoRPCAddress      = ctx.String(fCXORPCAddress)
		discoveryAddresses = ctx.StringSlice(fDiscoveryAddresses)

		rpcAddress  = ctx.String(fRPCAddress)
		remoteClose = ctx.Bool(fRemoteClose)
	)

	var (
		e        error
		stateDB  iko.StateDB
		cxoChain *iko.CXOChain
	)

	// Prepare StateDB.
	stateDB = iko.NewMemoryState()

	// Prepare ChainDB.
	cxoChain, e = iko.NewCXOChain(&iko.CXOChainConfig{
		Dir:                cxoDir,
		Public:             true,
		Memory:             testMode,
		MessengerAddresses: discoveryAddresses,
		CXOAddress:         cxoAddress,
		CXORPCAddress:      cxoRPCAddress,
		MasterRooter:       true,
		MasterRootPK:       rootPK,
		MasterRootSK:       rootSK,
		MasterRootNonce:    rootNonce,
	})
	if e != nil {
		return e
	}
	defer cxoChain.Close()

	// Prepare blockchain config.
	bcConfig := &iko.BlockChainConfig{
		GenerationPK: txGenPK,
		TransferPKs:  txTransPKs,
		TxAction: func(tx *iko.Transaction) error {
			// TODO: Implement web-socket.
			return nil
		},
	}

	// Prepare blockchain.
	bc, e := iko.NewBlockChain(bcConfig, cxoChain, stateDB)
	if e != nil {
		return e
	}
	defer bc.Close()

	if cxoChain != nil {
		cxoChain.RunTxService(iko.MakeTxChecker(bc))
	}

	if doInit || testMode {
		if e := cxoChain.MasterInitChain(); e != nil {
			return e
		}
	}

	log.Info("finished preparing blockchain")

	// Prepare test data.
	if testMode {
		var tx *iko.Transaction
		for i := 0; i < testCount; i++ {
			tx = iko.NewGenTx(iko.KittyID(i), testSK)

			log.WithField("tx", tx.String()).
				Debugf("test:tx_inject(%d)", i)

			if _, e := bc.InjectTx(tx); e != nil {
				return e
			}
		}
	}

	if testMode {
		tempDir, e := ioutil.TempDir(os.TempDir(), "kc")
		if e != nil {
			return e
		}
		defer os.RemoveAll(tempDir)
	}

	// Prepare rpc server.
	rpcServer, e := rpc.NewServer(
		&rpc.ServerConfig{
			Address:          rpcAddress,
			EnableRemoteQuit: remoteClose,
		},
		&rpc.Gateway{
			IKO:      bc,
			QuitChan: quit,
		},
	)
	if e != nil {
		return e
	}
	defer rpcServer.Close()

	<-quit
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
