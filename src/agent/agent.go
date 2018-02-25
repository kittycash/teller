// Package agent handles the kitty box reservations
package agent

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"github.com/kittycash/teller/src/util/httputil"
	"github.com/kittycash/teller/src/util/logger"
)

//@TODO proper error handling
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
	DoReservation(userAddress string, kittyID string, coinType string) error
	CancelReservation(kittyID string) error
	GetReservations(status string) ([]Reservation, error)
}

// Agent represents an agent object
type Agent struct {
	log   logrus.FieldLogger
	store Storer
	cfg   Config
}

// New creates a new agent service
func New(log logrus.FieldLogger, cfg Config, store Storer) *Agent {
	return &Agent{
		log: log.WithField("prefix", "teller.agent"),
		cfg: cfg,
		store: store,
	}
}

