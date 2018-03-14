package agent

import (
	"errors"
	"sync"
)

var (
	// ErrUserNotFound represents that the user does not exist
	ErrUserNotFound = errors.New("User not found")
)

// User represent a kitty cash user
type User struct {
	*sync.Mutex `json:"-"`
	// Users skycoin address
	Address string `json:"address"`
	// A user can have multiple reservations
	// capped by maxReservation
	Reservations []Reservation `json:"reservations"`
}

// CanReserve checks if the user can make any more reservations
func (u *User) CanReserve() bool {
	return len(u.Reservations) < maxReservation
}

// UserManager keeps tracks for user reservations
type UserManager struct {
	Users map[string]*User
}

// GetUser returns a user from the usermanager
func (um *UserManager) GetUser(userAddr string) (*User, error) {
	u, ok := um.Users[userAddr]
	if !ok {
		return nil, ErrUserNotFound
	}

	return u, nil
}

// AddReservation adds reservation to a usermanager
func (um *UserManager) AddReservation(userAddr string, reservation *Reservation) error {
	// Get the user and check whether he can reserve boxes
	u, err := um.GetUser(userAddr)
	if err != nil {
		return err
	}

	u.Lock()
	defer u.Unlock()
	if !u.CanReserve() {
		return ErrMaxReservationsExceeded
	}

	reservation.MakeReserved()
	u.Reservations = append(u.Reservations, *reservation)

	return nil
}
