package agent

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/util/dbutil"
)

var (
	// ReservationsKittyBkt bucket maps kitty id to reservation
	// Reservations will always exist against a kittyid
	ReservationsKittyBkt = []byte("reservations_kitty_idx")
	// UsersBkt maps user address to user info like reservations
	UsersBkt = []byte("users")
	// KittyOwnerBkt maps kitty id to the skycoin address of the user who has reserved it
	KittyOwnerBkt = []byte("kitty_owner_idex")
	// BoxBkt map kittyID to a kitty box
	// We do need some place to maintain kitty boxes
	// @TODO need to think more about this. What will we store and the structure
	BoxBkt = []byte("boxes")
)

// Storer interface handles database interactions
type Storer interface {
	GetReservations() ([]Reservation, error)
	GetReservationsByStatus(status string) ([]Reservation, error)
	GetReservationFromKittyID(kittyID string) (*Reservation, error)
	GetReservationUserFromKittyID(kittyID string) (*User, error)
	AddUser(user *User) error
	GetUsers() ([]User, error)
	GetUser(userAddr string) (*User, error)
	GetUserReservations(userAddr string) ([]Reservation, error)
	UpdateUser(user *User) error
	UpdateReservation(reservation *Reservation) error
	UpdateReservations(reservations []*Reservation) error
}

// Store saves reservations and user data
type Store struct {
	db  *bolt.DB
	log logrus.FieldLogger
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

// GetReservations fetches all the reservations from the database
func (s *Store) GetReservations() ([]Reservation, error) {
	var reservations []Reservation

	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		return dbutil.ForEach(tx, ReservationsKittyBkt, func(k, v []byte) error {
			var reservation Reservation

			err = json.Unmarshal(v, &reservation)
			if err != nil {
				return err
			}

			reservations = append(reservations, reservation)

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetReservationFromKittyID returns a reservation from the kittyID
// Args:
// kittyID: ID of the kitty in the reservation box
func (s *Store) GetReservationFromKittyID(kittyID string) (*Reservation, error) {
	reservation := &Reservation{}

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.GetBucketObject(tx, ReservationsKittyBkt, kittyID, reservation)
	}); err != nil {
		return nil, err
	}

	return reservation, nil
}

// GetReservationUserFromKittyID returns the user who has reserved the reservation for the given kittyID
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

// GetReservationsByStatus gets reversation based on the reservation status
// Args:
// status: Reservation status, availabled or reserved
func (s *Store) GetReservationsByStatus(status string) ([]Reservation, error) {
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

// Adduser adds a new user to the database
func (s *Store) AddUser(user *User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return dbutil.PutBucketValue(tx, UsersBkt, user.Address, user)
	})
}

// GetUsers fetchs all the users from the database
func (s *Store) GetUsers() ([]User, error) {
	var users []User

	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		return dbutil.ForEach(tx, UsersBkt, func(k, v []byte) error {
			var user User

			err = json.Unmarshal(v, &user)
			if err != nil {
				return err
			}

			users = append(users, user)

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return users, nil
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

// UpdateUser updates user info
// Args:
// User: object of the user to be updated
func (s *Store) UpdateUser(user *User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return dbutil.PutBucketValue(tx, UsersBkt, user.Address, user)
	})
}

// UpdateReservation Updates a reservation
// Args:
// reservation: reservation to be updated
func (s *Store) UpdateReservation(reservation *Reservation) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return dbutil.PutBucketValue(tx, ReservationsKittyBkt, reservation.KittyID, *reservation)
	})
}

// UpdateReservations Updates a list of reservations
// Args:
// reservations: reservations to be updated
func (s *Store) UpdateReservations(reservations []*Reservation) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var err error
		for _, r := range reservations {
			err = dbutil.PutBucketValue(tx, ReservationsKittyBkt, r.KittyID, *r)
			if err != nil {
				break
			}
		}
		return err
	})
}
