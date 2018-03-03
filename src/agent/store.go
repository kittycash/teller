package agent

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"errors"
	"github.com/kittycash/teller/src/util/dbutil"
	"github.com/kittycash/teller/src/box"
	"encoding/json"
)

var (
	// ReservationsKittyBkt bucket maps kitty id to reservation
	// Reservations will always exist against a kittyid
	ReservationsKittyBkt = []byte("reservations_kitty_idx")
	// Users bucket maps user address to user info like reservations
	UsersBkt = []byte("users")
	// KittyOwnerBkt maps kitty id to the skycoin address of the user who has reserved it
	KittyOwnerBkt = []byte("kitty_owner_idex")
	// Box bucket map kittyID to a kitty box
	// We do need some place to maintain kitty boxes
	// @TODO need to think more about this. What will we store and the structure
	BoxBkt = []byte("boxes")
)

// Storer interface handles database interactions
type Storer interface {
	GetReservations(status string) ([]Reservation, error)
	GetReservationFromKittyID(kittyID string) (*Reservation, error)
	GetReservationUserFromKittyID(kittyID string) (*User, error)
	GetUser(userAddr string) (*User, error)
	GetUserReservations(userAddr string) ([]Reservation, error)
	UpdateUser(user *User) error
	UpdateReservation(kittyID string, reservation *Reservation) error
}

// Store saves reservations and user data
type Store struct {
	db        *bolt.DB
	log       logrus.FieldLogger
}

// NewStore creates an agent Store
func NewStore(log logrus.FieldLogger, db *bolt.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("new Store failed: db is nil")
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		// create reservations bucket if not exist
		if _, err := tx.CreateBucketIfNotExists(ReservationsKittyBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(ReservationsKittyBkt, err)
		}

		// create users bucket if not exist
		if _, err := tx.CreateBucketIfNotExists(UsersBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(UsersBkt, err)
		}

		// create kitty owner bkt if not exist
		if _, err := tx.CreateBucketIfNotExists(KittyOwnerBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(KittyOwnerBkt, err)
		}
		// create box bucket if not exist
		if _, err := tx.CreateBucketIfNotExists(BoxBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(BoxBkt, err)
		}

		return nil
	}); err != nil {
		return nil, err
	}


	return &Store{
		db:  db,
		log: log,
	}, nil
}

// GetReservation returns a reservation instance
// Args:
// kittyID: ID of the kitty in the reservation box
func (s *Store) GetReservationFromKittyID(kittyID string) (*Reservation, error) {
	var reservation *Reservation

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.GetBucketObject(tx, ReservationsKittyBkt, kittyID, reservation)
	}); err != nil {
		return nil, err
	}

	return reservation, nil
}

// GetReservationUserFromKittyID
// Args:
// kittyID: ID of a kitty reserved by the user
func (s *Store) GetReservationUserFromKittyID(kittyID string) (*User, error) {
	var userAddr string

	// fetch the user address from kitty id
	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		userAddr, err = dbutil.GetBucketString(tx, KittyOwnerBkt, kittyID)
		return err
	}); err != nil {
		return nil, err
	}

	// fetch user from user address
	return s.GetUser(userAddr)
}

// GetReservations gets reversation based on the reservation status
// Args:
// status: Reservation status, availabled or reserved
func (s *Store) GetReservations(status string) ([]Reservation, error) {
	var reservations []Reservation

	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		return dbutil.ForEach(tx, ReservationsKittyBkt, func(k, v []byte) error {
			var reservation Reservation

			err = json.Unmarshal(v, &reservation)
			if err != nil {
				return err
			}

			// find the reservations with required status
			// append them to the reservations array
			if reservation.Status == status {
				reservations = append(reservations, reservation)
			}

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetUser gets user info from the user address
func (s *Store) GetUser(userAddr string) (*User, error) {
	var user *User

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.GetBucketObject(tx, UsersBkt, userAddr, user)
	}); err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserReservations gets user reservations from user address
func (s *Store) GetUserReservations(userAddr string) ([]Reservation, error) {
	user, err := s.GetUser(userAddr)
	if err != nil {
		return nil, err
	}

	return user.Reservations, nil
}

// Update user updates user info
// Args:
// User: object of the user to be updated
func (s *Store) UpdateUser(user *User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return dbutil.PutBucketValue(tx, UsersBkt, user.Address, user)
	})
}

// Updates a reservation
// Args:
// kittyID: ID of kitty in the reservation
// reservation: object of reservation to be updated
func (s *Store) UpdateReservation(kittyID string, reservation *Reservation) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return dbutil.PutBucketValue(tx, ReservationsKittyBkt, kittyID, reservation)
	})
}