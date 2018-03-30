// Package agent handles the kitty box reservations
package agent

import (
	"strconv"

	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/kitty-api/src/rpc"
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
	KittyAPIAddress string
}

// Manager provides APIs to interact with the agent service
type Manager interface {
	MakeReservation(userAddress string, kittyID string, coinType string, verificationCode string) error
	CancelReservation(userAddress, kittyID string) error
	GetReservations(status string) ([]Reservation, error)
	GetReservation(kittyID string) (*Reservation, error)
	GetKittyDepositAddress(kittyID string) (string, error)
}

// Agent represents an agent object
// handles kitty reservation requests
// enforces limits, verifies verification codes
type Agent struct {
	log                logrus.FieldLogger
	store              Storer
	cfg                Config
	ReservationManager *ReservationManager
	UserManager        *UserManager
	Verifier           *Verifier
	KittyAPI           *KittyAPIClient
}

// New creates a new agent service
func New(log logrus.FieldLogger, cfg Config, store Storer) *Agent {
	var um UserManager
	var rm ReservationManager

	verifier := NewVerifier(log)
	kittyAPICLient := NewKittyAPI(&rpc.ClientConfig{
		Address: cfg.KittyAPIAddress,
	}, log)

	// get 100 kitties from the start
	// no filters or sorters
	entries, err := kittyAPICLient.c.Entries(&rpc.EntriesIn{
		Offset:   0,
		PageSize: 100,
		Filters:  &database.Filters{},
		Sorters:  &database.Sorters{},
	})
	// panic if we are not able to fetch kitties from kitty api
	if err != nil {
		log.Panic(err)
	}

	for _, entry := range entries.Results {
		kittyID := strconv.Itoa(int(entry.ID))
		// panic if we come across a faulty kittyid
		if err != nil {
			log.Panic(err)
		}
		rm.Reservations[kittyID] = &Reservation{
			KittyID: kittyID,
			Status:  entry.Reservation,
		}
	}

	users, _ := store.GetUsers()
	for _, u := range users {
		um.Users[u.Address] = &u
	}

	return &Agent{
		log:                log.WithField("prefix", "teller.agent"),
		cfg:                cfg,
		store:              store,
		ReservationManager: &rm,
		UserManager:        &um,
		Verifier:           verifier,
		KittyAPI:           kittyAPICLient,
	}
}
