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
	mux sync.Mutex `json:"-"`
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

// UserManager keeps tracks of user reservations
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

// GetUser returns a user from the usermanager
func (um *UserManager) AddUser(user *User) {
	um.Users[user.Address] = user
}

// AddReservation adds reservation to a usermanager
func (um *UserManager) AddReservation(u *User, reservation *Reservation) error {
	u.mux.Lock()
	defer u.mux.Unlock()
	if !u.CanReserve() {
		return ErrMaxReservationsExceeded
	}

	reservation.MakeReserved()
	u.Reservations = append(u.Reservations, *reservation)

	return nil
}
