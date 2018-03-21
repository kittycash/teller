package agent

import (
	"sync"
	"time"

	"github.com/go-errors/errors"
)

//TODO (therealssj): handle reservation expiry

var (
	// ErrMaxReservationsExceeded represents that the user crossed his reservation limit
	ErrMaxReservationsExceeded = errors.New("User has exceeded the max number of reservations")
	// ErrBoxAlreadyReserved represents that the box has already been reserved
	ErrBoxAlreadyReserved = errors.New("Box already reserved")
	// ErrInvalidCoinType represents that the box is being reserved using an unsupported cointype
	ErrInvalidCoinType = errors.New("Invalid coin type")
	// ErrReservationNotFound reservation was not found in the database
	ErrReservationNotFound = errors.New("Reservation not found")
	// ErrInvalidReservationType invalid reservation type ( reserved / available )
	ErrInvalidReservationType = errors.New("Invalid reservation type")
	// ErrDepositAddressNotFound deposit address not found for the reservation
	ErrDepositAddressNotFound = errors.New("Deposit Address not found")
)

const (
	// Available reservation
	// newly added reversation or when a reservation expires is set to available
	Available = "available"
	// Reserved reservation
	Reserved = "reserved"
	// Delivered means the box of this reservation has been sent to a user
	Delivered = "delivered"
)

// Reservation is a reservation instance for a kitty box
type Reservation struct {
	// DepositAddress is where the buyer should send the payment
	DepositAddress string
	// address where we will send the kitty
	OwnerAddress string
	// KittyID is the unique ID of the kitty inside the box being reserved
	KittyID string
	// Status of the reservation
	Status string
	// Payment currency
	CoinType string
	// Expire defines after when a reservation expires
	Expire time.Time
}

// ReservationManager keeps track of reservations in the iko
type ReservationManager struct {
	*sync.RWMutex
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

// GetReservationByKittyID returns reservation of the kittyID
func (rm *ReservationManager) GetReservationByKittyID(kittyID string) (*Reservation, error) {
	rm.RLock()
	defer rm.RUnlock()

	// check if the reservation exists
	if _, ok := rm.Reservations[kittyID]; !ok {
		return nil, ErrReservationNotFound
	}

	return rm.Reservations[kittyID], nil
}

// GetReservations returns all reservations currently being tracked by reservation manager
func (rm *ReservationManager) GetReservations() []Reservation {
	rm.RLock()
	defer rm.RUnlock()

	var reservations []Reservation
	for _, r := range rm.Reservations {
		reservations = append(reservations, *r)
	}

	return reservations
}

// GetReservationsByStatus returns all reservations of given status
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

// ChangeReservationStatus changes status of a reservation
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
func (a *Agent) MakeReservation(userAddr string, kittyID string, cointype string, verificationCode string) error {
	// verify the verification code
	err := a.Verifier.VerifyCode(verificationCode)
	if err != nil {
		a.log.WithError(err).Error("Verifier.VerifyCode failed")
		return err
	}

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

	u, err := a.UserManager.GetUser(userAddr)
	if err != nil {
		a.log.WithError(err).Error("UserManager.GetUser failed")
		return err
	}

	// update the reservation
	if err := a.store.UpdateReservation(reservation.KittyID, reservation); err != nil {
		a.log.WithError(err).Errorf("CancelReservation failed for %s", reservation.KittyID)
		return err
	}

	// update the user
	err = a.store.UpdateUser(u)
	if err != nil {
		a.log.WithError(err).Error("Storer.UpdateUser failed")
		return err
	}
	
	// satisfy the verification code
	return a.Verifier.SatisfyCode(verificationCode)
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
		if user.Reservations[i].KittyID == kittyID {
			reservation = &user.Reservations[i]
			// make the reservation available
			reservation.MakeAvailable()
			a.ReservationManager.ChangeReservationStatus(kittyID, Available)
			// update the reservation
			if err := a.store.UpdateReservation(reservation.KittyID, reservation); err != nil {
				a.log.WithError(err).Errorf("CancelReservation failed for %s", reservation.KittyID)
				return err
			}

			// delete the reservation
			user.Reservations = append(user.Reservations[:i], user.Reservations[i+1:]...)

			if err := a.store.UpdateUser(user); err != nil {
				return err
			}
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
	case Available, Reserved, Delivered:
		return a.ReservationManager.GetReservationsByStatus(status), nil
	case "all":
		return a.ReservationManager.GetReservations(), nil
	default:
		return nil, ErrInvalidReservationType
	}
}

// GetReservation gets reversation from the kitty id
// Args:
// status: kittyID
func (a *Agent) GetReservation(kittyID string) (*Reservation, error) {
	return a.store.GetReservationFromKittyID(kittyID)
}

// GetKittyDepositAddress gets deposit address of kitty box reservation
// Args:
// kittyID: ID of kitty inside the box
func (a *Agent) GetKittyDepositAddress(kittyID string) (string, error) {
	reservation, err := a.store.GetReservationFromKittyID(kittyID)
	if err != nil {
		a.log.WithError(err).Errorf("GetKittyDepositAddress failed for %v", kittyID)
		return "", err
	}

	if reservation.DepositAddress == "" {
		return "", ErrDepositAddressNotFound
	}

	return reservation.DepositAddress, nil
}
