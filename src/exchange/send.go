package exchange

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/config"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/wallet/src/iko"
)

//@TODO needs to be refactored for sending boxes

// Sender is a component for sending boxes
type Sender interface {
	Status() error
	Balance() int
}

// SendRunner a Sender than can be run
type SendRunner interface {
	Runner
	Sender
}

// Send reads deposits from a Processor and sends coins
type Send struct {
	log         logrus.FieldLogger
	cfg         config.BoxExchanger
	processor   Processor
	sender      sender.Sender // sender provides APIs for sending skycoin
	store       Storer        // deposit info storage
	quit        chan struct{}
	done        chan struct{}
	depositChan chan DepositInfo
	statusLock  sync.RWMutex
	status      error
}

// NewSend creates exchange service
func NewSend(log logrus.FieldLogger, cfg config.BoxExchanger, store Storer, sender sender.Sender, processor Processor) (*Send, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if cfg.TxConfirmationCheckWait == 0 {
		cfg.TxConfirmationCheckWait = txConfirmationCheckWait
	}

	return &Send{
		cfg:         cfg,
		log:         log.WithField("prefix", "teller.exchange.send"),
		processor:   processor,
		sender:      sender,
		store:       store,
		quit:        make(chan struct{}),
		done:        make(chan struct{}, 1),
		depositChan: make(chan DepositInfo, 100),
	}, nil
}

// Run starts the exchange process
func (s *Send) Run() error {
	log := s.log
	log.Info("Start exchange service...")
	defer func() {
		log.Info("Closed exchange service")
		s.done <- struct{}{}
	}()

	var wg sync.WaitGroup

	if s.cfg.SendEnabled {
		// Load StatusWaitSend deposits for processing later
		waitSendDeposits, err := s.store.GetDepositInfoArray(func(di DepositInfo) bool {
			return di.Status == StatusWaitSend
		})

		if err != nil {
			err = fmt.Errorf("GetDepositInfoArray failed: %v", err)
			log.WithError(err).Error(err)
			return err
		}

		// Load StatusWaitConfirm deposits for processing later
		waitConfirmDeposits, err := s.store.GetDepositInfoArray(func(di DepositInfo) bool {
			return di.Status == StatusWaitConfirm
		})

		if err != nil {
			err = fmt.Errorf("GetDepositInfoArray failed: %v", err)
			log.WithError(err).Error(err)
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.runSend()
		}()

		// Queue the saved StatusWaitConfirm deposits
		for _, di := range waitConfirmDeposits {
			s.depositChan <- di
		}

		// Queue the saved StatusWaitSend deposits
		for _, di := range waitSendDeposits {
			s.depositChan <- di
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.runNoSend()
		}()
	}

	// Merge processor.Deposits() into the internal depositChan
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.receiveDeposits()
	}()

	wg.Wait()

	return nil
}

func (s *Send) runSend() {
	// This loop processes StatusWaitSend deposits.
	// Only one deposit is processed at a time; it will not send more coins
	// until it receives confirmation of the previous send.
	log := s.log.WithField("goroutine", "runSend")
	for {
		select {
		case <-s.quit:
			log.Info("quit")
			return
		case d := <-s.depositChan:
			log := log.WithField("depositInfo", d)
			if err := s.processWaitSendDeposit(d); err != nil {
				log.WithError(err).Error("processWaitSendDeposit failed. This deposit will not be reprocessed until teller is restarted.")
			}
		}
	}
}

func (s *Send) runNoSend() {
	// Flush the deposit channel so that it doesn't fill up
	log := s.log.WithField("goroutine", "runNoSend")
	for {
		select {
		case <-s.quit:
			log.Info("quit")
			return
		case d := <-s.depositChan:
			log := log.WithField("depositInfo", d)
			log.Warning("Received depositInfo, but sending is disabled")
		}
	}
}

func (s *Send) receiveDeposits() {
	// Read deposits from the processor and place them on the internal deposit channel
	// This is necessary as a separate step because deposits can come from other sources,
	// specifically deposits that were partially in processing from a previous run will
	// be loaded first, and not come from the receiver.
	log := s.log.WithField("goroutine", "receiveDeposits")
	for {
		select {
		case <-s.quit:
			log.Info("quit")
			return
		case d := <-s.processor.Deposits():
			log.WithField("depositInfo", d).Info("Received deposit from processor")
			s.depositChan <- d
		}
	}
}

// Shutdown close the exchange service
func (s *Send) Shutdown() {
	close(s.quit)
	s.log.Info("Waiting for Run() to finish")
	<-s.done
	s.log.Info("Shutdown complete")
}

// processDeposit advances a single deposit through three states:
// StatusWaitSend -> StatusWaitConfirm
// StatusWaitConfirm -> StatusDone
// StatusWaitDeposit is never saved to the database, so it does not transition
func (s *Send) processWaitSendDeposit(di DepositInfo) error {
	log := s.log.WithField("depositInfo", di)
	log.Info("Processing StatusWaitSend deposit")

	for {
		select {
		case <-s.quit:
			return nil
		default:
		}

		log.Info("handleDepositInfoState")

		var err error
		di, err = s.handleDepositInfoState(di)
		log = log.WithField("depositInfo", di)

		s.setStatus(err)

		switch err.(type) {
		case sender.APIError:
			// Treat kitty client errors as temporary.
			// Some API errors are hypothetically permanent,
			// but most likely it is an insufficient wallet balance or
			// the kitty node is unavailable.
			// A permanent error suggests a bug in kittycash or teller so can be fixed.
			log.WithError(err).Error("handleDepositInfoState failed")
			select {
			case <-time.After(s.cfg.TxConfirmationCheckWait):
			case <-s.quit:
				return nil
			}
		default:
			switch err {
			case nil:
				break
			case ErrNotConfirmed:
				select {
				case <-time.After(s.cfg.TxConfirmationCheckWait):
				case <-s.quit:
					return nil
				}
			default:
				log.WithError(err).Error("handleDepositInfoState failed")
				return err
			}
		}

		if di.Status == StatusDone {
			return nil
		}
	}
}

func (s *Send) handleDepositInfoState(di DepositInfo) (DepositInfo, error) {
	log := s.log.WithField("deposit", di)

	if err := di.ValidateForStatus(); err != nil {
		log.WithError(err).Error("handleDepositInfoState's DepositInfo is invalid")
		return di, err
	}

	switch di.Status {
	case StatusWaitSend:
		//@TODO (therealssj): implement

		return di, nil

	case StatusWaitConfirm:
		// Wait for confirmation
		rsp := s.sender.IsTxConfirmed(di.TxHash)

		if rsp == nil {
			log.WithError(ErrNoResponse).Warn("Sender closed")
			return di, ErrNoResponse
		}

		if rsp.Err != nil {
			log.WithError(rsp.Err).Error("IsTxConfirmed failed")
			return di, rsp.Err
		}

		if !rsp.Confirmed {
			log.Info("Transaction is not confirmed yet")
			return di, ErrNotConfirmed
		}

		log.Info("Transaction is confirmed")

		di, err := s.store.UpdateDepositInfo(di.DepositID, func(di DepositInfo) DepositInfo {
			di.Status = StatusDone
			return di
		})
		if err != nil {
			log.WithError(err).Error("UpdateDepositInfo set StatusDone failed")
			return di, err
		}

		log.Info("DepositInfo status set to StatusDone")

		return di, nil

	case StatusDone:
		log.Warn("DepositInfo already processed")
		return di, nil

	case StatusWaitDeposit:
		// We don't save any deposits with StatusWaitDeposit.
		// We can't transition to StatusWaitSend without a scanner.Deposit
		log.Error("StatusWaitDeposit cannot be processed and should never be handled by this method")
		fallthrough
	case StatusUnknown:
		fallthrough
	default:
		err := ErrDepositStatusInvalid
		log.WithError(err).Error(err)
		return di, err
	}
}

func (s *Send) createTransaction(di DepositInfo) (*iko.Transaction, error) {
	log := s.log.WithField("deposit", di)

	// This should never occur, the DepositInfo is saved with a DepositAddress
	// during GetOrCreateDepositInfo().
	if di.DepositAddress == "" {
		err := ErrNoBoundAddress
		log.WithError(err).Error(err)
		return nil, err
	}

	log = log.WithField("depositAddress", di.DepositAddress)
	log = log.WithField("ownerAddress", di.OwnerAddress)
	log = log.WithField("kittyID", di.KittyID)
	log = log.WithField("depositAmt", di.DepositValue)

	//@TODO (therealssj): verify deposit amount here

	log.Info("Creating kitty cash transaction")

	kID, err := iko.KittyIDFromString(di.KittyID)
	if err != nil {
		log.WithError(err).Errorf("failed to convert kittyID %v", di.KittyID)
		return nil, err
	}

	tx, err := s.sender.CreateTransaction(di.OwnerAddress, kID)
	if err != nil {
		log.WithError(err).Error("sender.CreateTransaction failed")
		return nil, err
	}

	//@TODO (therealssj): do transaction verification here

	return tx, nil
}

func verifyCreatedTransaction(tx *iko.Transaction, di DepositInfo) error {
	// Check invariant assertions:
	// The transaction should contain one output to the destination address.
	// It may or may not have a change output.
	//@TODO (therealssj): implement

	return nil
}

func (s *Send) broadcastTransaction(tx *iko.Transaction) (*sender.BroadcastTxResponse, error) {
	log := s.log.WithField("txid", tx.Hash().Hex())

	log.Info("Broadcasting skycoin transaction")

	rsp := s.sender.BroadcastTransaction(tx)

	log = log.WithField("sendRsp", rsp)

	if rsp == nil {
		err := ErrNoResponse
		log.WithError(err).Warn("Sender closed")
		return nil, err
	}

	if rsp.Err != nil {
		err := fmt.Errorf("Send skycoin failed: %v", rsp.Err)
		log.WithError(err).Error(err)
		return nil, err
	}

	log.Info("Sent skycoin")

	return rsp, nil
}

// Balance is broken right now
func (s *Send) Balance() int {
	return s.sender.Balance()
}

func (s *Send) setStatus(err error) {
	defer s.statusLock.Unlock()
	s.statusLock.Lock()
	s.status = err
}

// Status returns the last return value of the processing state
func (s *Send) Status() error {
	defer s.statusLock.RUnlock()
	s.statusLock.RLock()
	return s.status
}
