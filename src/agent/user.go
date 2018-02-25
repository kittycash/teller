package agent


// User represent a kitty cash user
type User struct {
	// Users skycoin address
	Address string
	// A user can have multiple reservations
	// capped by maxReservation
	Reservations []Reservation
}

// CanReserve checks if the user can make any more reservations
func (u *User) CanReserve() bool {
	return len(u.Reservations) < maxReservation
}

// AddReservation updates a reservation and assigns it a user
func (u *User) AddReservation(reservation *Reservation) {
	// change the reservation status to reserved
	reservation.MakeReserved()
	//@TODO
	// set deposit address
	u.Reservations = append(u.Reservations, *reservation)
}
