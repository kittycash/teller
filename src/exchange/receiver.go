package exchange

import (
	"fmt"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/config"
	"github.com/kittycash/teller/src/scanner"
)

//func init() {
//	cfg := config.BoxExchanger{	}
//}

// Receiver is a component that reads deposits from a scanner.Scanner and records them
type Receiver interface {
	Deposits() <-chan DepositInfo
	BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error)
	BindAddressWithTx(tx *bolt.Tx, kittyID, depositAddr, coinType string) (*BoundAddress, error)
}

// ReceiveRunner is a Receiver than can be run
type ReceiveRunner interface {
	Runner
	Receiver
}

// Receive implements a Receiver. All incoming deposits are saved,
// with the configured rate recorded at instantiation time [TODO: move that functionality to Processor?]
type Receive struct {
	log         logrus.FieldLogger
	cfg         config.BoxExchanger
	multiplexer *scanner.Multiplexer
	store       Storer
	deposits    chan DepositInfo
	quit        chan struct{}
	done        chan struct{}
}

// NewReceive creates a Receive
func NewReceive(log logrus.FieldLogger, cfg config.BoxExchanger, store Storer, multiplexer *scanner.Multiplexer) (*Receive, error) {
	// TODO -- split up config into relevant parts?
	// The Receive component needs exchange rates
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Receive{
		log:         log.WithField("prefix", "teller.exchange.Receive"),
		cfg:         cfg,
		store:       store,
		multiplexer: multiplexer,
		deposits:    make(chan DepositInfo, 100),
		quit:        make(chan struct{}),
		done:        make(chan struct{}),
	}, nil
}

// Run processes deposits from the scanner.Scanner, recording them and exposing them over the Deposits() channel
func (r *Receive) Run() error {
	log := r.log
	log.Info("Start receive service...")
	defer func() {
		log.Info("Closed receive service")
		r.done <- struct{}{}
	}()

	// Load StatusWaitDecide deposits for resubmission
	waitDecideDeposits, err := r.store.GetDepositInfoArray(func(di DepositInfo) bool {
		return di.Status == StatusWaitDecide
	})

	if err != nil {
		err = fmt.Errorf("GetDepositInfoArray failed: %v", err)
		log.WithError(err).Error(err)
		return err
	}

	// Queue the saved StatusWaitDecide deposits
	// This will block if there are too many waiting deposits, make sure that
	// the Processor is running to receive them
	for _, di := range waitDecideDeposits {
		r.deposits <- di
	}

	var wg sync.WaitGroup

	// This loop processes incoming deposits from the scanner and saves a
	// new DepositInfo with a status of StatusWaitSend
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.runReadMultiplexer()
	}()

	wg.Wait()

	return nil
}

// runReadMultiplexer reads deposits from the multiplexer
func (r *Receive) runReadMultiplexer() {
	log := r.log.WithField("goroutine", "readMultiplexer")
	for {
		var dv scanner.DepositNote
		var ok bool
		select {
		case <-r.quit:
			log.Info("quit")
			return
		case dv, ok = <-r.multiplexer.GetDeposit():
			if !ok {
				log.Warn("Scan service closed, watch deposits loop quit")
				return
			}

		}
		log := log.WithField("deposit", dv.Deposit)

		// Save a new DepositInfo based upon the scanner.Deposit.
		// If the save fails, report it to the scanner.
		// The scanner will mark the deposit as "processed" if no error
		// occurred.  Any unprocessed deposits held by the scanner
		// will be resent to the exchange when teller is started.
		if d, err := r.saveIncomingDeposit(dv.Deposit); err != nil {
			log.WithError(err).Error("saveIncomingDeposit failed. This deposit will not be reprocessed until teller is restarted.")
			dv.ErrC <- err
		} else {
			dv.ErrC <- nil
			r.deposits <- d
		}
	}
}

// Shutdown stops a previous call to run
func (r *Receive) Shutdown() {
	r.log.Info("Shutting down Receive")
	close(r.quit)
	r.log.Info("Waiting for run to finish")
	<-r.done
	r.log.Info("Shutdown complete")
}

// Deposits returns a channel with recorded deposits
func (r *Receive) Deposits() <-chan DepositInfo {
	return r.deposits
}

// saveIncomingDeposit is called when receiving a deposit from the scanner
func (r *Receive) saveIncomingDeposit(dv scanner.Deposit) (DepositInfo, error) {
	log := r.log.WithField("deposit", dv)

	di, err := r.store.GetOrCreateDepositInfo(dv)
	if err != nil {
		log.WithError(err).Error("GetOrCreateDepositInfo failed")
		return DepositInfo{}, err
	}

	log = log.WithField("depositInfo", di)
	log.Info("Saved DepositInfo")

	return di, err
}

// BindAddress binds deposit address with kitty id, and
// add the btc/sky address to scan service, when a deposit is detected
// to the btc/sky address, will send specific kitty box to the user who owns the box
func (r *Receive) BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	if err := r.multiplexer.ValidateCoinType(coinType); err != nil {
		return nil, err
	}

	boundAddr, err := r.store.BindAddress(kittyID, depositAddr, coinType)
	if err != nil {
		return nil, err
	}

	if err := r.multiplexer.AddScanAddress(depositAddr, coinType); err != nil {
		return nil, err
	}

	return boundAddr, nil
}

func (r *Receive) BindAddressWithTx(tx *bolt.Tx, kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	if err := r.multiplexer.ValidateCoinType(coinType); err != nil {
		return nil, err
	}

	boundAddr, err := r.store.BindAddressWithTx(tx, kittyID, depositAddr, coinType)
	if err != nil {
		return nil, err
	}

	if err := r.multiplexer.AddScanAddress(depositAddr, coinType); err != nil {
		return nil, err
	}

	return boundAddr, nil
}
