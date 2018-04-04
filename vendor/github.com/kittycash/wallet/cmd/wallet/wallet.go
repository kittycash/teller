package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/util/file"
	"gopkg.in/sirupsen/logrus.v1"
	"gopkg.in/urfave/cli.v1"

	"github.com/kittycash/wallet/src/http"
	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/util"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/kittycash/wallet/src/kitties"
)

const (
	// TODO: Define proper values for these!
	TrustedGenPK     = "03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753"
	TrustedRootPK    = "03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753"
	TrustedRootNonce = uint64(79)
	TrustedAPIDomain = "https://api.kittycash.com"

	DefaultHttpAddress = "127.0.0.1:7908"
	DefaultCXOAddress  = "127.0.0.1:7900"
	DefaultDiscovery   = ""

	DirRoot         = ".kittycash"
	DirChildCXO     = "cxo"
	DirChildWallets = "wallets"
)

const (
	fCXODir             = "cxo-dir"
	fCXOAddress         = "cxo-address"
	fCXORPCAddress      = "cxo-rpc-address"
	fDiscoveryAddresses = "messenger-addresses"

	fWalletDir = "wallet-dir"

	fHttpAddress = "http-address"
	fGUI         = "gui"
	fGUIDir      = "gui-dir"
	fTLS         = "tls"
	fTLSCert     = "tls-cert"
	fTLSKey      = "tls-key"

	fTest          = "test"
	fTestGenPK     = "test-gen-pk"
	fTestRootPK    = "test-root-pk"
	fTestRootNonce = "test-root-nonce"
	fTestAPIDomain = "test-api-domain"
)

func Flag(flag string, short ...string) string {
	if len(short) == 0 {
		return flag
	}
	return flag + ", " + short[0]
}

var (
	app       = cli.NewApp()
	log       = logrus.New()
	homeDir   = file.UserHome()
	staticDir = func() string {
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			return filepath.Join(goPath, "src/github.com/kittycash/wallet/wallet/dist")
		}
		return "./static/dist"
	}()
)

func init() {
	app.Name = "wallet"
	app.Description = "kitty cash wallet executable"
	app.Flags = cli.FlagsByName{
		/*
			<<< CXO CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(fCXODir),
			Usage: "directory to store cxo files",
			Value: filepath.Join(homeDir, DirRoot, DirChildCXO),
		},
		cli.StringFlag{
			Name:  Flag(fCXOAddress),
			Usage: "address to use to serve CXO",
			Value: DefaultCXOAddress,
		},
		cli.StringSliceFlag{
			Name:  Flag(fDiscoveryAddresses),
			Usage: "discovery addresses",
			Value: &cli.StringSlice{DefaultDiscovery},
		},
		cli.StringFlag{
			Name:  Flag(fCXORPCAddress),
			Usage: "address for CXO RPC, leave blank to disable CXO RPC",
		},
		/*
			<<< WALLET CONFIG >>>
		*/
		cli.StringFlag{
			Name:  Flag(fWalletDir),
			Usage: "directory to store wallet files",
			Value: filepath.Join(homeDir, DirRoot, DirChildWallets),
		},
		/*
			<<< HTTP SERVER >>>
		*/
		cli.StringFlag{
			Name:  Flag(fHttpAddress),
			Usage: "address to serve http server on",
			Value: DefaultHttpAddress,
		},
		cli.BoolTFlag{
			Name:  Flag(fGUI),
			Usage: "whether to enable gui",
		},
		cli.StringFlag{
			Name:  Flag(fGUIDir),
			Usage: "directory to serve GUI from",
			Value: staticDir,
		},
		cli.BoolFlag{
			Name:  Flag(fTLS),
			Usage: "whether to enable tls",
		},
		cli.StringFlag{
			Name:  Flag(fTLSCert),
			Usage: "tls certificate file path",
		},
		cli.StringFlag{
			Name:  Flag(fTLSKey),
			Usage: "tls key file path",
		},
		/*
			<<< TEST MODE >>>
		*/
		cli.BoolFlag{
			Name:  Flag(fTest),
			Usage: "whether to run wallet in test mode",
		},
		cli.StringFlag{
			Name:  Flag(fTestRootPK),
			Usage: "test mode trusted root public key",
		},
		cli.Uint64Flag{
			Name:  Flag(fTestRootNonce),
			Usage: "test mode trusted root nonce",
		},
		cli.StringFlag{
			Name:  Flag(fTestGenPK),
			Usage: "test mode trusted gen tx public key",
		},
		cli.StringFlag{
			Name: Flag(fTestAPIDomain),
			Usage: "test mode kitty-api domain to use",
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := util.CatchInterrupt()

	var (
		rootPK = cipher.MustPubKeyFromHex(TrustedRootPK)
		rootNc = TrustedRootNonce
		genPK  = cipher.MustPubKeyFromHex(TrustedGenPK)
		apiDomain = TrustedAPIDomain

		walletDir = ctx.String(fWalletDir)

		cxoDir             = ctx.String(fCXODir)
		cxoAddress         = ctx.String(fCXOAddress)
		cxoRPCAddress      = ctx.String(fCXORPCAddress)
		discoveryAddresses = ctx.StringSlice(fDiscoveryAddresses)

		httpAddress = ctx.String(fHttpAddress)
		gui         = ctx.BoolT(fGUI)
		guiDir      = ctx.String(fGUIDir)
		tls         = ctx.Bool(fTLS)
		tlsCert     = ctx.String(fTLSCert)
		tlsKey      = ctx.String(fTLSKey)

		test = ctx.Bool(fTest)
	)

	// Test mode changes.
	if test {
		rootPK = cipher.MustPubKeyFromHex(ctx.String(fTestRootPK))
		rootNc = ctx.Uint64(fTestRootNonce)
		genPK = cipher.MustPubKeyFromHex(ctx.String(fTestGenPK))
		apiDomain = ctx.String(fTestAPIDomain)

		tempDir, err := ioutil.TempDir(os.TempDir(), "kc_wallet")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		walletDir = tempDir
	}

	// Prepare StateDB.
	stateDB := iko.NewMemoryState()

	// Prepare ChainDB.
	cxoChain, err := iko.NewCXOChain(&iko.CXOChainConfig{
		Dir:                cxoDir,
		Public:             true,
		Memory:             test,
		MessengerAddresses: discoveryAddresses,
		CXOAddress:         cxoAddress,
		CXORPCAddress:      cxoRPCAddress,
		MasterRooter:       false,
		MasterRootPK:       rootPK,
		MasterRootNonce:    rootNc,
	})
	if err != nil {
		return err
	}
	defer cxoChain.Close()

	// Prepare blockchain config.
	bcConfig := &iko.BlockChainConfig{
		GenerationPK: genPK,
		TxAction: func(tx *iko.Transaction) error {
			return nil
		},
	}

	// Prepare blockchain.
	bc, err := iko.NewBlockChain(bcConfig, cxoChain, stateDB)
	if err != nil {
		return err
	}
	defer bc.Close()

	if cxoChain != nil {
		cxoChain.RunTxService(iko.MakeTxChecker(bc, true))
	}

	log.Info("finished preparing blockchain")

	// Prepare wallet.
	if err := wallet.SetRootDir(walletDir); err != nil {
		return err
	}
	walletManager, err := wallet.NewManager()
	if err != nil {
		return err
	}

	// Prepare market.
	market, err := kitties.NewManager(&kitties.ManagerConfig{
		KittyAPIDomain: apiDomain,
	})
	if err != nil {
		return err
	}

	// Prepare http server.
	httpServer, err := http.NewServer(
		&http.ServerConfig{
			Address:     httpAddress,
			EnableGUI:   gui,
			GUIDir:      guiDir,
			EnableTLS:   tls,
			TLSCertFile: tlsCert,
			TLSKeyFile:  tlsKey,
		},
		&http.Gateway{
			IKO:    bc,
			Wallet: walletManager,
			Market: market,
		},
	)
	if err != nil {
		return err
	}
	defer httpServer.Close()

	<-quit
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}
