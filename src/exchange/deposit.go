package exchange

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kittycash/teller/src/scanner"
)

// Status deposit Status
type Status int8

const (
	// StatusWaitDeposit deposit has not occurred
	StatusWaitDeposit Status = iota
	// StatusWaitSend deposit is ready for send
	StatusWaitSend
	// StatusWaitConfirm kitty sent, but not confirmed yet
	StatusWaitConfirm
	// StatusDone kitty sent and confirmed
	StatusDone
	// StatusUnknown fallback value
	StatusUnknown
	// StatusWaitDecide initial deposit receive state
	StatusWaitDecide
)

var statusString = []string{
	StatusWaitDeposit: "waiting_deposit",
	StatusWaitSend:    "waiting_send",
	StatusWaitConfirm: "waiting_confirm",
	StatusDone:        "done",
	StatusUnknown:     "unknown",
	StatusWaitDecide:  "waiting_decide",
}

func (s Status) String() string {
	return statusString[s]
}

// NewStatusFromStr create status from string
func NewStatusFromStr(st string) Status {
	switch st {
	case statusString[StatusWaitDeposit]:
		return StatusWaitDeposit
	case statusString[StatusWaitSend]:
		return StatusWaitSend
	case statusString[StatusWaitConfirm]:
		return StatusWaitConfirm
	case statusString[StatusDone]:
		return StatusDone
	case statusString[StatusWaitDecide]:
		return StatusWaitDecide
	default:
		return StatusUnknown
	}
}

// BoundAddress records information about an address binding
type BoundAddress struct {
	KittyID  string
	Address  string
	CoinType string
}

// DepositInfo records the deposit info
type DepositInfo struct {
	Seq            uint64
	UpdatedAt      int64
	Status         Status
	CoinType       string
	KittyID        string
	DepositAddress string
	OwnerAddress   string
	DepositID      string
	Txid           string // txhash
	DepositValue   int64  // Deposit amount. Should be measured in the smallest unit possible (e.g. satoshis for BTC or droplets for skycoin)
	Error          string // An error that occurred during processing
	// The original Deposit is saved for the records, in case there is a mistake.
	// Do not use this data directly.  All necessary data is copied to the top level
	// of DepositInfo (e.g. DepositID, DepositAddress, DepositValue, CoinType).
	Deposit scanner.Deposit
}

// DepositStats records overall statistics about deposits
type DepositStats struct {
	TotalBTCReceived int64 `json:"total_btc_received"`
	TotalSKYReceived int64 `json:"total_sky_received"`
	TotalBoxesSent   int64 `json:"total_boxes_sent"`
}

// ValidateForStatus does a consistency check of the data based upon the Status value
func (di DepositInfo) ValidateForStatus() error {

	checkWaitSend := func() error {
		if di.Seq == 0 {
			return errors.New("Seq missing")
		}
		if di.KittyID == "" {
			return errors.New("KittyID missing")
		}
		if di.DepositAddress == "" {
			return errors.New("DepositAddress missing")
		}
		if di.DepositID == "" {
			return errors.New("DepositID missing")
		}
		if di.CoinType == scanner.CoinTypeBTC && !isValidBtcTx(di.DepositID) {
			return fmt.Errorf("Invalid DepositID value \"%s\"", di.DepositID)
		}
		if di.DepositValue == 0 {
			return errors.New("DepositValue is zero")
		}

		return nil
	}

	switch di.Status {
	case StatusDone:
		if di.Error != ErrEmptySendAmount.Error() && di.Txid == "" {
			return errors.New("Txid missing")
		}
		// Don't check SkySent == 0, it is possible to have StatusDone with
		// no sky sent, this happens if we received a very tiny deposit.
		// In that case, the DepositInfo status is switched to StatusDone
		// without doing StatusWaitConfirm
		return checkWaitSend()

	case StatusWaitConfirm:
		if di.Txid == "" {
			return errors.New("Txid missing")
		}

		return checkWaitSend()

	case StatusWaitSend:
		return checkWaitSend()

	case StatusWaitDecide:
		return checkWaitSend()

	case StatusWaitDeposit, StatusUnknown:
		fallthrough
	default:
		return fmt.Errorf("DepositInfo should not have status %s[%d]", di.Status.String(), di.Status)
	}
}

func isValidBtcTx(btcTx string) bool {
	if btcTx == "" {
		return false
	}

	pts := strings.Split(btcTx, ":")
	if len(pts) != 2 {
		return false
	}

	if pts[0] == "" || pts[1] == "" {
		return false
	}

	_, err := strconv.ParseInt(pts[1], 10, 64)
	return err == nil
}
