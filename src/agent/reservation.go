package agent

import (
	"github.com/kittycash/teller/src/box"
	"time"
	"github.com/go-errors/errors"
)

// @TODO handle reservation expiry
// @TODO prevent reservation by two people at the same time

var (
	ErrMaxReservationsExceeded = errors.New("User has exceeded the max number of reservations")
	ErrBoxAlreadyReserved = errors.New("Box already reserved")
	ErrInvalidCoinType = errors.New("Invalid coin type")
	ErrReservationNotFound = errors.New("Reservation not found")
	ErrInvalidReservationType = errors.New("Invalid reservation type")
)

const (
	// Available reservation
	Available  = "available"
	// Reserved reservation
	Reserved  = "reserved"
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

// MakeReserved marks a reservation as reserved
func (r *Reservation) MakeReserved()  {
	r.Status = Reserved
}

// MakeAvailable marks a reservation as available
func (r *Reservation) MakeAvailable() {
	r.Status = Available
}


// DoReservation reserves a kitty box
// Args:
// userAddress: Address of the user reserving the box
// kittyID: ID of kitty in the reservation box
// cointype: payment cointype
func (a *Agent) DoReservation(user string, kittyID string, cointype string) error {
	// Create a user instance and check whether he can reserve boxes
	u, err := a.store.GetUser(user)
	if err != nil {
		a.log.WithError(err).Error("store.GetUser failed")
		return err
	}

	if !u.CanReserve() {
		return ErrMaxReservationsExceeded
	}

	// fetch the kitty reservation
	kr, err := a.store.GetReservationFromKittyID(kittyID)
	if err != nil {
		a.log.WithError(err).Error("store.GetReservationFromKittyID failed")
		return err
	}

	// check whether the kitty is available or not
	switch kr.Status {
	case Reserved:
		return ErrBoxAlreadyReserved
	case Available:
		// set the payment cointype
		switch cointype {
		case "SKY":
		case "BTC":
			kr.CoinType = cointype
		default:
			return ErrInvalidCoinType
		}
	}

	u.AddReservation(kr)
	// update the reservation
	a.store.UpdateReservation(kr.Box.KittyID, kr)
	// update the user
	a.store.UpdateUser(u)

	return nil
}


// CancelReservation cancels a kitty reservation
// Args:
// userAddress: Address of the user reserving the box
// kittyID: ID of kitty in the reservation box
func (a *Agent) CancelReservation(kittyID string) error {
	// fetch the user who has reserved the kitty box
	user, err := a.store.GetReservationUserFromKittyID(kittyID)
	if err != nil {
		a.log.WithError(err).Error("GetReservationUserFromKittyID failed")
		return err
	}
	// get the reservations of the user
	reservations, err := a.store.GetUserReservations(user.Address)
	if err != nil {
		a.log.WithError(err).Error("GetUserReservations failed")
		return err
	}

	var reservation *Reservation
	for i := range reservations {
		if reservations[i].Box.KittyID == kittyID {
			reservation = &reservations[i]
			// make the reservation available
			reservation.MakeAvailable()
			// update the reservation
			if err := a.store.UpdateReservation(reservation.Box.KittyID, reservation); err != nil {
				a.log.WithError(err).Error("CancelReservation failed for %s", reservation.Box.KittyID)
				return err
			}

			// delete the reservation
			reservations = append(reservations[:i], reservations[i+1:]...)

			// update user reservations
			user.Reservations = reservations
			a.store.UpdateUser(user)
		}
	}

	if reservation == nil {
		return ErrReservationNotFound
	}

	return nil
}

// GetReservations gets reversation based on the reservation status
// Args:
// status: Reservation status, availabled or reserved
func (a *Agent) GetReservations(status string) ([]Reservation, error) {
	switch status {
	case Available, Reserved:
		return a.store.GetReservations(status)
	default:
		return nil, ErrInvalidReservationType
	}
}