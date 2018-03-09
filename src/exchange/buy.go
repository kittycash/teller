package exchange

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/config"
)

// Processor is a component that processes deposits from a Receiver and sends them to a Sender
type Processor interface {
	Deposits() <-chan DepositInfo
}

// ProcessRunner is a Processor that can be run
type ProcessRunner interface {
	Runner
	Processor
}

// Buy implements a Processor. All deposits are sent directly to the sender for processing.
type Buy struct {
	log      logrus.FieldLogger
	cfg      config.BoxExchanger
	receiver Receiver
	store    Storer
	deposits chan DepositInfo
	quit     chan struct{}
	done     chan struct{}
}

// NewBuy creates DirectBuy
func NewBuy(log logrus.FieldLogger, cfg config.BoxExchanger, store Storer, receiver Receiver) (*Buy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Buy{
		log:      log.WithField("prefix", "teller.exchange.directbuy"),
		cfg:      cfg,
		store:    store,
		receiver: receiver,
		deposits: make(chan DepositInfo, 100),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
	}, nil
}

// Run updates all deposits with StatusWaitSend and exposes them over Deposits()
func (p *Buy) Run() error {
	log := p.log
	log.Info("Start direct buy service...")
	defer func() {
		log.Info("Closed direct buy service")
		p.done <- struct{}{}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.runUpdateStatus()
	}()

	wg.Wait()

	return nil
}

// runUpdateStatus reads deposits from the Receiver and changes their status to StatusWaitSend
func (p *Buy) runUpdateStatus() {
	log := p.log.WithField("goroutine", "runUpdateStatus")
	for {
		select {
		case <-p.quit:
			log.Info("quit")
			return
		case d := <-p.receiver.Deposits():
			updatedDeposit, err := p.updateStatus(d)
			if err != nil {
				msg := "runUpdateStatus failed. This deposit will not be reprocessed until teller is restarted."
				log.WithField("depositInfo", d).WithError(err).Error(msg)
				continue
			}

			//TODO (therealssj): make this resumable, needs to happen in a single transaction
			p.checkDepositProgress(&updatedDeposit)

			if updatedDeposit.Status == StatusWaitSend {
				p.deposits <- updatedDeposit
			}
		}
	}
}

// Shutdown stops a previous call to Run
func (p *Buy) Shutdown() {
	p.log.Info("Shutting down DirectBuy")
	close(p.quit)
	p.log.Info("Waiting for run to finish")
	<-p.done
	p.log.Info("Shutdown complete")
}

// Deposits returns a channel of processed deposits
func (p *Buy) Deposits() <-chan DepositInfo {
	return p.deposits
}

// updateStatus sets the deposit's status to StatusWaitPartial.
func (p *Buy) updateStatus(di DepositInfo) (DepositInfo, error) {
	updatedDi, err := p.store.UpdateDepositInfo(di.DepositID, func(di DepositInfo) DepositInfo {
		di.Status = StatusWaitPartial
		return di
	})
	if err != nil {
		p.log.WithError(err).Error("UpdateDepositInfo set StatusWaitPartial failed")
		return di, err
	}

	return updatedDi, nil
}

// checkDepositProgress checks the progress towards the full payment of the box
func (p *Buy) checkDepositProgress(di *DepositInfo) {
	dt, _ := p.store.getDepositTrack(di.DepositAddress)

	if dt.AmountDeposited+di.DepositValue >= dt.AmountRequired {
		dt.AmountDeposited += di.DepositValue
		p.store.updateDepositTrack(di.DepositAddress, dt)
		di.Status = StatusWaitSend
	}
}
