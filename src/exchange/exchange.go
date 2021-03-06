package exchange

import (
	"errors"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/config"
	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/sender"
)

const (
	txConfirmationCheckWait = time.Second * 3
)

var (
	// ErrEmptySendAmount is returned if we try to send no kitty
	ErrEmptySendAmount = errors.New("Sending no kitty")
	// ErrNoResponse is returned when the send service returns a nil response. This happens if the send service has closed.
	ErrNoResponse = errors.New("No response from the send service")
	// ErrNotConfirmed is returned if the tx is not confirmed yet
	ErrNotConfirmed = errors.New("Transaction is not confirmed yet")
	// ErrDepositStatusInvalid is returned when handling a deposit with a status that cannot be processed
	// This includes StatusWaitDeposit and StatusUnknown
	ErrDepositStatusInvalid = errors.New("Deposit status cannot be handled")
	// ErrNoBoundAddress is returned if no skycoin address is bound to a deposit's address
	ErrNoBoundAddress = errors.New("Deposit has no bound skycoin address")
)

// DepositFilter filters deposits
type DepositFilter func(di DepositInfo) bool

// Runner defines an interface for components that can be started and stopped
type Runner interface {
	Run() error
	Shutdown()
}

// Exchanger provides APIs to interact with the exchange service
type Exchanger interface {
	BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error)
	BindAddressWithTx(tx *bolt.Tx, kittyID, depositAddr, coinType string) (*BoundAddress, error)
	GetDepositStatuses(kittyID string) ([]DepositStatus, error)
	GetDepositStatusDetail(flt DepositFilter) ([]DepositStatusDetail, error)
	IsBound(kittyAddr string) bool
	GetDepositStats() (*DepositStats, error)
	Status() error
	Balance() (int, error)
}

// Exchange encompasses an entire coin<>skycoin deposit-process-send flow
type Exchange struct {
	log   logrus.FieldLogger
	store Storer
	cfg   config.BoxExchanger
	quit  chan struct{}
	done  chan struct{}

	Receiver  ReceiveRunner
	Processor ProcessRunner
	Sender    SendRunner
}

// NewExchange creates an Exchange which performs handles payments and forwards to sender once the payment is confirmed
func NewExchange(log logrus.FieldLogger, cfg config.BoxExchanger, store Storer, multiplexer *scanner.Multiplexer, boxSender sender.Sender) (*Exchange, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	receiver, err := NewReceive(log, cfg, store, multiplexer)
	if err != nil {
		return nil, err
	}

	processor, err := NewBuy(log, cfg, store, receiver)
	if err != nil {
		return nil, err
	}

	sender, err := NewSend(log, cfg, store, boxSender, processor)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		log:       log.WithField("prefix", "teller.exchange.exchange"),
		store:     store,
		cfg:       cfg,
		quit:      make(chan struct{}),
		done:      make(chan struct{}),
		Receiver:  receiver,
		Processor: processor,
		Sender:    sender,
	}, nil
}

// Run runs all components of the Exchange
func (e *Exchange) Run() error {
	e.log.Info("Start exchange service...")
	defer func() {
		e.log.Info("Closed exchange service")
		e.done <- struct{}{}
	}()

	// TODO: Alternative way of managing the subcomponents:
	// Create channels for linking two components, initialize the components with the channels
	// Close them to teardown

	errC := make(chan error, 3)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := e.Receiver.Run(); err != nil {
			e.log.WithError(err).Error("Receiver.Run failed")
			errC <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := e.Processor.Run(); err != nil {
			e.log.WithError(err).Error("Processor.Run failed")
			errC <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := e.Sender.Run(); err != nil {
			e.log.WithError(err).Error("Sender.Run failed")
			errC <- err
		}
	}()

	var err error
	select {
	case <-e.quit:
	case err = <-errC:
		e.log.WithError(err).Error("Terminating early")
	}

	wg.Wait()

	return err
}

// Shutdown stops a previous call to run
func (e *Exchange) Shutdown() {
	e.log.Info("Shutting down Exchange")
	close(e.quit)

	e.log.Info("Shutting down Exchange subcomponents")
	e.Receiver.Shutdown()
	e.Processor.Shutdown()
	e.Sender.Shutdown()

	e.log.Info("Waiting for run to finish")
	<-e.done
	e.log.Info("Shutdown complete")
}

// DepositStatus json struct for deposit status
type DepositStatus struct {
	Seq       uint64 `json:"seq"`
	UpdatedAt int64  `json:"updated_at"`
	Status    string `json:"status"`
	CoinType  string `json:"coin_type"`
}

// DepositStatusDetail deposit status detail info
type DepositStatusDetail struct {
	Seq            uint64 `json:"seq"`
	UpdatedAt      int64  `json:"updated_at"`
	Status         string `json:"status"`
	KittyID        string `json:"kitty_id"`
	DepositAddress string `json:"deposit_address"`
	OwnerAddress   string `json:"owner_address"`
	CoinType       string `json:"coin_type"`
	Txid           string `json:"txid"`
}

// GetDepositStatuses returns deamon.DepositStatus array of given skycoin address
func (e *Exchange) GetDepositStatuses(kittyID string) ([]DepositStatus, error) {
	dis, err := e.store.GetDepositInfoOfKittyID(kittyID)
	if err != nil {
		return []DepositStatus{}, err
	}

	dss := make([]DepositStatus, 0, len(dis))
	for _, di := range dis {
		dss = append(dss, DepositStatus{
			Seq:       di.Seq,
			UpdatedAt: di.UpdatedAt,
			Status:    di.Status.String(),
			CoinType:  di.CoinType,
		})
	}
	return dss, nil
}

// GetDepositStatusDetail returns deposit status details
func (e *Exchange) GetDepositStatusDetail(flt DepositFilter) ([]DepositStatusDetail, error) {
	dis, err := e.store.GetDepositInfoArray(flt)
	if err != nil {
		return nil, err
	}

	dss := make([]DepositStatusDetail, 0, len(dis))
	for _, di := range dis {
		dss = append(dss, DepositStatusDetail{
			Seq:            di.Seq,
			UpdatedAt:      di.UpdatedAt,
			Status:         di.Status.String(),
			KittyID:        di.KittyID,
			DepositAddress: di.DepositAddress,
			Txid:           di.Txid,
			CoinType:       di.CoinType,
			OwnerAddress:   di.OwnerAddress,
		})
	}
	return dss, nil
}

// IsBound returns whether the kitty is already bound to a deposit address or not
func (e *Exchange) IsBound(kittyID string) bool {
	//@TODO: improve this
	addr, _ := e.store.GetKittyBindAddress(kittyID)
	return addr != nil
}

// GetDepositStats returns deposit status
func (e *Exchange) GetDepositStats() (*DepositStats, error) {
	tbr, tsr, tbs, err := e.store.GetDepositStats()
	if err != nil {
		return nil, err
	}

	return &DepositStats{
		TotalBTCReceived: tbr,
		TotalSKYReceived: tsr,
		TotalBoxesSent:   tbs,
	}, nil
}

// Balance returns the number of coins left in the OTC wallet
func (e *Exchange) Balance() (int, error) {
	return e.Sender.Balance()
}

// Status returns the last return value of the processing state
func (e *Exchange) Status() error {
	return e.Sender.Status()
}

// BindAddress binds deposit address with kitty id of a box, and
// add the btc/sky address to scan service, when a deposit is detected
// to the btc/sky address, will send specific kitty box to the user who owns the box
func (e *Exchange) BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	return e.Receiver.BindAddress(kittyID, depositAddr, coinType)
}

func (e *Exchange) BindAddressWithTx(tx *bolt.Tx, kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	return e.Receiver.BindAddressWithTx(tx, kittyID, depositAddr, coinType)
}
