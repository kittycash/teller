// Skycoin teller, which provides service of monitoring the bitcoin deposite
// and sending skycoin coins
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/google/gops/agent"
	"github.com/skycoin/skycoin/src/cipher"

	"github.com/boltdb/bolt"
	btcrpcclient "github.com/btcsuite/btcd/rpcclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/kittycash/teller/src/addrs"
	kittyagent "github.com/kittycash/teller/src/agent"
	"github.com/kittycash/teller/src/config"
	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/monitor"
	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/teller/src/teller"
	"github.com/kittycash/teller/src/util/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createBtcScanner(log logrus.FieldLogger, cfg config.Config, scanStore *scanner.Store) (*scanner.BTCScanner, error) {
	// create btc rpc client
	certs, err := ioutil.ReadFile(cfg.BtcRPC.Cert)
	if err != nil {
		return nil, fmt.Errorf("Failed to read cfg.BtcRPC.Cert %s: %v", cfg.BtcRPC.Cert, err)
	}

	log.Info("Connecting to btcd")

	btcrpc, err := btcrpcclient.New(&btcrpcclient.ConnConfig{
		Endpoint:     "ws",
		Host:         cfg.BtcRPC.Server,
		User:         cfg.BtcRPC.User,
		Pass:         cfg.BtcRPC.Pass,
		Certificates: certs,
	}, nil)
	if err != nil {
		log.WithError(err).Error("Connect btcd failed")
		return nil, err
	}

	log.Info("Connect to btcd succeeded")

	err = scanStore.AddSupportedCoin(scanner.CoinTypeBTC)
	if err != nil {
		log.WithError(err).Error("scanStore.AddSupportedCoin(scanner.CoinTypeBTC) failed")
		return nil, err
	}

	btcScanner, err := scanner.NewBTCScanner(log, scanStore, btcrpc, scanner.Config{
		ScanPeriod:            cfg.BtcScanner.ScanPeriod,
		ConfirmationsRequired: cfg.BtcScanner.ConfirmationsRequired,
		InitialScanHeight:     cfg.BtcScanner.InitialScanHeight,
	})
	if err != nil {
		log.WithError(err).Error("Open scan service failed")
		return nil, err
	}
	return btcScanner, nil
}

// createSkyScanner returns a new sky scanner instance
func createSkyScanner(log logrus.FieldLogger, cfg config.Config, scanStore *scanner.Store) (*scanner.SKYScanner, error) {
	skyrpc := scanner.NewSkyClient(cfg.SkyRPC.Address)
	err := scanStore.AddSupportedCoin(scanner.CoinTypeSKY)
	if err != nil {
		log.WithError(err).Error("scanStore.AddSupportedCoin(scanner.CoinTypeSKY) failed")
	}

	skyScanner, err := scanner.NewSKYScanner(log, scanStore, skyrpc, scanner.Config{
		ScanPeriod:        cfg.SkyScanner.ScanPeriod,
		InitialScanHeight: cfg.SkyScanner.InitialScanHeight,
	})
	if err != nil {
		log.WithError(err).Error("Open skyscan service failed")
		return nil, err
	}

	return skyScanner, nil
}

func run() error {
	cur, err := user.Current()
	if err != nil {
		fmt.Println("Failed to get user's home directory:", err)
		return err
	}
	defaultAppDir := filepath.Join(cur.HomeDir, ".teller-kittycash")

	appDirOpt := pflag.StringP("dir", "d", defaultAppDir, "application data directory")
	configNameOpt := pflag.StringP("config", "c", "config", "name of configuration file")
	pflag.Parse()

	if err := createFolderIfNotExist(*appDirOpt); err != nil {
		fmt.Println("Create application data directory failed:", err)
		return err
	}

	cfg, err := config.Load(*configNameOpt, *appDirOpt)
	if err != nil {
		return fmt.Errorf("Config error:\n%v", err)
	}

	// Init logger
	rusloggger, err := logger.NewLogger(cfg.LogFilename, cfg.Debug)
	if err != nil {
		fmt.Println("Failed to create Logrus logger:", err)
		return err
	}

	log := rusloggger.WithField("prefix", "teller")

	log.WithField("config", cfg.Redacted()).Info("Loaded teller config")

	if cfg.Profile {
		// Start gops agent, for profiling
		if err := agent.Listen(agent.Options{
			ShutdownCleanup: true,
		}); err != nil {
			log.WithError(err).Error("Start profile agent failed")
			return err
		}
	}

	quit := make(chan struct{})
	go catchInterrupt(quit)

	// Open db
	dbPath := filepath.Join(*appDirOpt, cfg.DBFilename)
	db, err := bolt.Open(dbPath, 0700, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.WithError(err).Error("Open db failed")
		return err
	}

	errC := make(chan error, 20)
	var wg sync.WaitGroup

	background := func(name string, errC chan<- error, f func() error) {
		log.Infof("Backgrounding task %s", name)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := f()
			if err != nil {
				log.WithError(err).Errorf("Backgrounded task %s failed", name)
				errC <- fmt.Errorf("Backgrounded task %s failed: %v", name, err)
			} else {
				log.Infof("Backgrounded task %s shutdown", name)
			}
		}()
	}

	var btcScanner *scanner.BTCScanner
	var skyScanner *scanner.SKYScanner
	var scanService scanner.Scanner
	var scanSkyService scanner.Scanner
	var sendService *sender.SendService
	var sendAPI sender.Sender
	var btcAddrMgr *addrs.Addrs
	var skyAddrMgr *addrs.Addrs

	//create multiplexer to manage scanner
	multiplexer := scanner.NewMultiplexer(log)

	dummyMux := http.NewServeMux()

	// create scan storer
	scanStore, err := scanner.NewStore(log, db)
	if err != nil {
		log.WithError(err).Error("scanner.NewStore failed")
		return err
	}

	if cfg.Dummy.Scanner {
		log.Info("btcd disabled, running dummy scanner")
		scanService = scanner.NewDummyScanner(log)
		scanService.(*scanner.DummyScanner).RegisterCoinType(scanner.CoinTypeBTC)
		// TODO -- refactor dummy scanning to support multiple coin types
		// scanEthService = scanner.NewDummyScanner(log)
		scanService.(*scanner.DummyScanner).BindHandlers(dummyMux)
	} else {
		// enable btc scanner
		if cfg.BtcScanner.Enabled {
			btcScanner, err = createBtcScanner(rusloggger, cfg, scanStore)
			if err != nil {
				log.WithError(err).Error("create btc scanner failed")
				return err
			}
			background("btcScanner.Run", errC, btcScanner.Run)

			scanService = btcScanner
		}

		// create sky scanner if its enabled
		if cfg.SkyScanner.Enabled {
			skyScanner, err = createSkyScanner(rusloggger, cfg, scanStore)
			if err != nil {
				log.WithError(err).Error("create sky scanner failed")
				return err
			}

			background("skyscanner.Run", errC, skyScanner.Run)

			scanSkyService = skyScanner

			if err := multiplexer.AddScanner(scanSkyService, scanner.CoinTypeSKY); err != nil {
				log.WithError(err).Errorf("multiplexer.AddScanner of %s failed", scanner.CoinTypeSKY)
				return err
			}
		}
	}

	if err := multiplexer.AddScanner(scanService, scanner.CoinTypeBTC); err != nil {
		log.WithError(err).Errorf("multiplexer.AddScanner of %s failed", scanner.CoinTypeBTC)
		return err
	}

	background("multiplex.Run", errC, multiplexer.Multiplex)

	if cfg.Dummy.Sender {
		log.Info("kittyd disabled, running dummy sender")
		//sendAPI = sender.NewDummySender(log)
		//sendAPI.(*sender.DummySender).BindHandlers(dummyMux)
	} else {
		kittyClient, err := sender.NewRPC(cfg.KittyClientAddr)
		if err != nil {
			log.WithError(err).Error("sender.NewAPI failed")
			return err
		}

		secKey := cipher.MustSecKeyFromHex(cfg.SecKey)
		sendService = sender.NewService(log, kittyClient, secKey)

		background("sendService.Run", errC, sendService.Run)

		sendAPI = sender.NewRetrySender(sendService)
	}

	if cfg.Dummy.Scanner || cfg.Dummy.Sender {
		log.Infof("Starting dummy admin interface listener on http://%s", cfg.Dummy.HTTPAddr)
		go func() {
			if err := http.ListenAndServe(cfg.Dummy.HTTPAddr, dummyMux); err != nil {
				log.WithError(err).Error("Dummy ListenAndServe failed")
			}
		}()
	}

	// create exchange service
	exchangeStore, err := exchange.NewStore(log, db)
	if err != nil {
		log.WithError(err).Error("exchange.NewStore failed")
		return err
	}

	exchangeClient, err := exchange.NewExchange(log, cfg.BoxExchanger, exchangeStore, multiplexer, sendAPI)
	if err != nil {
		log.WithError(err).Error("exchange.NewDirectExchange failed")
		return err
	}

	background("exchangeClient.Run", errC, exchangeClient.Run)

	//create AddrManager
	addrManager := addrs.NewAddrManager()

	if cfg.BtcScanner.Enabled {
		// create bitcoin address manager
		btcAddrMgr, err = addrs.NewBTCAddrs(log, db, cfg.BtcAddresses)
		if err != nil {
			log.WithError(err).Error("Create bitcoin deposit address manager failed")
			return err
		}
		if err := addrManager.PushGenerator(btcAddrMgr, scanner.CoinTypeBTC); err != nil {
			log.WithError(err).Error("Add BTC address manager failed")
			return err
		}
	}

	if cfg.SkyScanner.Enabled {
		// create sky address manager
		skyAddrMgr, err = addrs.NewSKYAddrs(log, db, cfg.SkyAddresses)
		if err != nil {
			log.WithError(err).Error("Create SKY deposit address manager failed")
			return err
		}
		if err := addrManager.PushGenerator(skyAddrMgr, scanner.CoinTypeSKY); err != nil {
			log.WithError(err).Error("Add SKY address manager failed")
			return err
		}
	}

	// create agent store
	agentStore, err := kittyagent.NewStore(log, db)
	if err != nil {
		log.WithError(err).Error("agent.NewStore failed")
		return err
	}

	// create a new agent manager instance
	agentCfg := kittyagent.Config{
		KittyAPIAddress: cfg.KittyApi.Address,
		VerifierEnabled: cfg.VerificationService.Enabled,
	}
	agentManager := kittyagent.New(log, agentCfg, agentStore)
	tellerServer := teller.New(log, exchangeClient, addrManager, agentManager, cfg, db)

	// Run the service
	background("tellerServer.Run", errC, tellerServer.Run)

	// start monitor service
	monitorCfg := monitor.Config{
		Addr: cfg.AdminPanel.Host,
	}
	monitorService := monitor.New(log, monitorCfg, btcAddrMgr, skyAddrMgr, exchangeClient, btcScanner)

	background("monitorService.Run", errC, monitorService.Run)

	var finalErr error
	select {
	case <-quit:
	case finalErr = <-errC:
		if finalErr != nil {
			log.WithError(finalErr).Error("Goroutine error")
		}
	}

	log.Info("Shutting down...")

	if monitorService != nil {
		log.Info("Shutting down monitorService")
		monitorService.Shutdown()
	}

	// close the teller service
	log.Info("Shutting down tellerServer")
	tellerServer.Shutdown()

	log.Info("Shutting down the multiplexer")
	multiplexer.Shutdown()

	// close the scan service
	if btcScanner != nil {
		log.Info("Shutting down btcScanner")
		btcScanner.Shutdown()
	}

	// close exchange service
	log.Info("Shutting down exchangeClient")
	exchangeClient.Shutdown()

	// close the skycoin send service
	if sendService != nil {
		log.Info("Shutting down sendService")
		sendService.Shutdown()
	}

	log.Info("Waiting for goroutines to exit")

	wg.Wait()

	log.Info("Shutdown complete")

	return finalErr
}

func createFolderIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// create the dir
		if err := os.Mkdir(path, 0700); err != nil {
			return err
		}
	}
	return nil
}

func printProgramStatus() {
	p := pprof.Lookup("goroutine")
	if err := p.WriteTo(os.Stdout, 2); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
}

func catchInterrupt(quit chan<- struct{}) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	signal.Stop(sigchan)
	close(quit)

	// If ctrl-c is called again, panic so that the program state can be examined.
	// Ctrl-c would be called again if program shutdown was stuck.
	go catchInterruptPanic()
}

// catchInterruptPanic catches os.Interrupt and panics
func catchInterruptPanic() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	signal.Stop(sigchan)
	printProgramStatus()
	panic("SIGINT")
}
