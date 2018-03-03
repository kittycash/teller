// Package agent handles the kitty box reservations
package agent

import (
	"github.com/sirupsen/logrus"
	"github.com/kittycash/wallet/src/iko"
)

//@TODO implement limits

const (
	// Max number of reservations a user can do
	maxReservation = 1
)

// Config defines the agent config
type Config struct {
	//@TODO
}

// AgentManager provides APIs to interact with the agent service
type AgentManager interface {
	MakeReservation(userAddress string, kittyID string, coinType string) error
	CancelReservation(kittyID string) error
	GetReservations(status string) ([]Reservation, error)
	GetKittyDepositAddress(kittyID string) (string, error)
}

// Agent represents an agent object
type Agent struct {
	log   logrus.FieldLogger
	store Storer
	cfg   Config
	ReservationManager *ReservationManager
	UserManager *UserManager
}

// New creates a new agent service
func New(log logrus.FieldLogger, cfg Config, store Storer) *Agent {
	return &Agent{
		log:   log.WithField("prefix", "teller.agent"),
		cfg:   cfg,
		store: store,
	}
}
