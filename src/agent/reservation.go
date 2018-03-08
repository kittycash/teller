package agent

import (
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/kittycash/teller/src/box"
)

// @TODO handle reservation expiry

var (
	ErrMaxReservationsExceeded = errors.New("User has exceeded the max number of reservations")
	ErrBoxAlreadyReserved      = errors.New("Box already reserved")
	ErrInvalidCoinType         = errors.New("Invalid coin type")
	ErrReservationNotFound     = errors.New("Reservation not found")
	ErrInvalidReservationType  = errors.New("Invalid reservation type")
	ErrDepositAddressNotFound  = errors.New("Deposit Address not found")
)

const (
	// Available reservation
	Available = "available"
	// Reserved reservation
	Reserved = "reserved"
)

// Reservation is a reservation instance for a kitty box
type Reservation struct {
	// Address where the buyer should send the payment
	DepositAddress string
	// The kitty box of this reservation
	Box *box.Box
	// Status of the reservation
	Status string
	// Payment currency
	CoinType string
	// Expire defines after when a reservation expires
	Expire time.Time
}

// Reservation keeps track of reservations in the iko
type ReservationManager struct {
	sync.RWMutex
	Reservations map[string]*Reservation
}

// MakeReserved marks a reservation as reserved
func (r *Reservation) MakeReserved() {
	r.Status = Reserved
}

// MakeAvailable marks a reservation as available
func (r *Reservation) MakeAvailable() {
	r.Status = Available
}

func (rm *ReservationManager) GetReservationByKittyID(kittyID string) (*Reservation, error) {
	rm.RLock()
	defer rm.RUnlock()

	// check if the reservation exists
	if _, ok := rm.Reservations[kittyID]; !ok {
		return nil, ErrReservationNotFound
	}

	return rm.Reservations[kittyID], nil
}

func (rm *ReservationManager) GetReservationsByStatus(status string) []Reservation {
	rm.RLock()
	defer rm.RUnlock()
	var reservations []Reservation

	for _, r := range rm.Reservations {
		if r.Status == status {
			reservations = append(reservations, *r)
		}
	}

	return reservations
}

func (rm *ReservationManager) GetReservations() []Reservation {
	rm.RLock()
	defer rm.RUnlock()

	var reservations []Reservation
	for _, r := range rm.Reservations {
		reservations = append(reservations, *r)
	}

	return reservations
}

func (rm *ReservationManager) ChangeReservationStatus(kittyID string, status string) {
	rm.Lock()
	defer rm.Unlock()
	rm.Reservations[kittyID].Status = status
}

// MakeReservation reserves a kitty box
// Args:
// userAddress: Address of the user reserving the box
// kittyID: ID of kitty in the reservation box
// cointype: payment cointype
func (a *Agent) MakeReservation(userAddr string, kittyID string, cointype string) error {
	// get the reservation for the reservation map
	reservation, err := a.ReservationManager.GetReservationByKittyID(kittyID)
	if err != nil {
		a.log.WithError(err).Error("ReservationManager.GetReservation failed")
		return err
	}

	// check whether the kitty is available or not
	switch reservation.Status {
	case Reserved:
		return ErrBoxAlreadyReserved
	case Available:
		// set the payment cointype
		switch cointype {
		case "SKY":
		case "BTC":
			reservation.CoinType = cointype
		default:
			return ErrInvalidCoinType
		}
	}

	// set the reservation as reserved
	a.ReservationManager.ChangeReservationStatus(kittyID, Reserved)

	err = a.UserManager.AddReservation(userAddr, reservation)
	if err != nil {
		a.log.WithError(err).Error("UserManager.AddReservation failed")
		return err
	}

	return nil
}

// CancelReservation cancels a kitty reservation
// Args:
// userAddress: Address of the user reserving the box
// kittyID: ID of kitty in the reservation box
func (a *Agent) CancelReservation(userAddress, kittyID string) error {
	user, err := a.UserManager.GetUser(userAddress)
	if err != nil {
		return err
	}
	var reservation *Reservation
	for i := range user.Reservations {
		if user.Reservations[i].Box.KittyID == kittyID {
			reservation = &user.Reservations[i]
			// make the reservation available
			reservation.MakeAvailable()
			a.ReservationManager.ChangeReservationStatus(kittyID, Available)
			// update the reservation
			if err := a.store.UpdateReservation(reservation.Box.KittyID, reservation); err != nil {
				a.log.WithError(err).Error("CancelReservation failed for %s", reservation.Box.KittyID)
				return err
			}

			// delete the reservation
			user.Reservations = append(user.Reservations[:i], user.Reservations[i+1:]...)

			a.store.UpdateUser(user)
			break
		}
	}

	if reservation == nil {
		return ErrReservationNotFound
	}

	return nil
}

// GetReservations gets reversation based on the reservation status
// Args:
// status: Reservation status, available, reserved or all.
func (a *Agent) GetReservations(status string) ([]Reservation, error) {
	switch status {
	case Available, Reserved:
		return a.ReservationManager.GetReservationsByStatus(status), nil
	case "all":
		return a.ReservationManager.GetReservations(), nil
	default:
		return nil, ErrInvalidReservationType
	}
}

// GetKittyDepositAddress gets deposit address of kitty box reservation
// Args:
// kittyID: ID of kitty inside the box
func (a *Agent) GetKittyDepositAddress(kittyID string) (string, error) {
	reservation, err := a.store.GetReservationFromKittyID(kittyID)
	if err != nil {
		a.log.WithError(err).Error("GetKittyDepositAddress failed for %v", kittyID)
		return "", err
	}

	if reservation.DepositAddress == "" {
		return "", ErrDepositAddressNotFound
	}

	return reservation.DepositAddress, nil
}
