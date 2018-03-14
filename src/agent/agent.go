// Package agent handles the kitty box reservations
package agent

import (
	"github.com/sirupsen/logrus"
)

//TODO (therealssj): implement limits
//TODO (therealssj): improve data structures

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
	CancelReservation(userAddress, kittyID string) error
	GetReservations(status string) ([]Reservation, error)
	GetReservation(kittyID string) (*Reservation, error)
	GetKittyDepositAddress(kittyID string) (string, error)
}

// Agent represents an agent object
type Agent struct {
	log                logrus.FieldLogger
	store              Storer
	cfg                Config
	ReservationManager *ReservationManager
	UserManager        *UserManager
}

// New creates a new agent service
func New(log logrus.FieldLogger, cfg Config, store Storer) *Agent {
	reservations, _ := store.GetReservations()
	users, _ := store.GetUsers()

	var um UserManager
	var rm ReservationManager

	for _, r := range reservations {
		rm.Reservations[r.Box.KittyID] = &r
	}

	for _, u := range users {
		um.Users[u.Address] = &u
	}
	return &Agent{
		log:                log.WithField("prefix", "teller.agent"),
		cfg:                cfg,
		store:              store,
		ReservationManager: &rm,
		UserManager:        &um,
	}
}
