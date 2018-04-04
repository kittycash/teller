// Package agent handles the kitty box reservations
package agent

import (
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/kittycash/kitty-api/src/rpc"
	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/util/dbutil"
)

const (
	// Max number of reservations a user can do
	maxReservation = 1
)

// Config defines the agent config
type Config struct {
	KittyAPIAddress string
	VerifierEnabled bool
}

// Manager provides APIs to interact with the agent service
type Manager interface {
	MakeReservation(tx *bolt.Tx, depositAddress, userAddress, kittyID, coinType, verificationCode string) error
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
	um := UserManager{
		Users: make(map[string]*User),
	}
	var rm ReservationManager
	verifier := NewVerifier(log, cfg.VerifierEnabled)
	kittyAPICLient := NewKittyAPI(&rpc.ClientConfig{
		Address: cfg.KittyAPIAddress,
	}, log)

	// get 100 kitties from the start
	// no filters or sorters
	entries, err := kittyAPICLient.c.Entries(&rpc.EntriesIn{
		Offset:   0,
		PageSize: 100,
	})
	//panic if we are not able to fetch kitties from kitty api
	if err != nil {
		log.Panic(err)
	}

	rm.Reservations = make(map[string]*Reservation, entries.TotalCount)
	for _, entry := range entries.Results {
		kittyID := strconv.Itoa(int(entry.ID))
		// panic if we come across a faulty kittyid
		if err != nil {
			log.Panic(err)
		}
		// fetch reservation from database to see if its exists or not
		r, err := store.GetReservationFromKittyID(kittyID)
		switch err.(type) {
		case dbutil.ObjectNotExistErr:
			rm.Reservations[kittyID] = &Reservation{
				KittyID:  kittyID,
				Status:   entry.Reservation,
				PriceBTC: entry.PriceBTC,
				PriceSKY: entry.PriceSKY,
			}
			err := store.UpdateReservation(rm.Reservations[kittyID])
			if err != nil {
				log.Panic(err)
			}
		case nil:
			// allow price to be controlled by the kitty api
			r.PriceSKY = entry.PriceSKY
			r.PriceBTC = entry.PriceBTC
			rm.Reservations[kittyID] = r
		default:
			log.Panic(err)
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
