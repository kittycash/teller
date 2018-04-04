package teller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kittycash/kitty-api/src/rpc"
	"github.com/kittycash/wallet/src/iko"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/cipher"

	"github.com/kittycash/teller/src/addrs"
	"github.com/kittycash/teller/src/agent"
	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/teller/src/util/httputil"
	"github.com/kittycash/teller/src/util/logger"
)

// StatusResponse http response for /api/status
type StatusResponse struct {
	Statuses []exchange.DepositStatus `json:"statuses,omitempty"`
}

// StatusHandler returns the deposit status of specific skycoin address
// Method: GET
// URI: /api/status
// Args:
//     skyaddr
func StatusHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		if !validMethod(ctx, w, r, []string{http.MethodGet}) {
			return
		}

		skyAddr := r.URL.Query().Get("skyaddr")

		// Remove extraneous whitespace
		skyAddr = strings.Trim(skyAddr, "\n\t ")

		if skyAddr == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing skyaddr"))
			return
		}

		log = log.WithField("skyAddr", skyAddr)
		ctx = logger.WithContext(ctx, log)

		log.Info()

		if !verifySkycoinAddress(ctx, w, skyAddr) {
			return
		}

		log.Info("Sending StatusRequest to teller")

		depositStatuses, err := s.service.GetDepositStatuses(skyAddr)
		if err != nil {
			log.WithError(err).Error("service.GetDepositStatuses failed")
			errorResponse(ctx, w, http.StatusInternalServerError, errInternalServerError)
			return
		}

		log = log.WithFields(logrus.Fields{
			"depositStatuses":    depositStatuses,
			"depositStatusesLen": len(depositStatuses),
		})
		log.Info("Got depositStatuses")

		if err := httputil.JSONResponse(w, StatusResponse{
			Statuses: depositStatuses,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

// ConfigResponse http response for /api/config
type ConfigResponse struct {
	Enabled                  bool  `json:"enabled"`
	BtcConfirmationsRequired int64 `json:"btc_confirmations_required"`
	MaxBoundAddresses        int   `json:"max_bound_addrs"`
	MaxDecimals              int   `json:"max_decimals"`
}

// ConfigHandler returns the teller configuration
// Method: GET
// URI: /api/config
func ConfigHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		if !validMethod(ctx, w, r, []string{http.MethodGet}) {
			return
		}

		if err := httputil.JSONResponse(w, ConfigResponse{
			Enabled:                  s.cfg.Teller.BindEnabled,
			MaxDecimals:              s.cfg.BoxExchanger.MaxDecimals,
			MaxBoundAddresses:        s.cfg.Teller.MaxBoundAddresses,
			BtcConfirmationsRequired: s.cfg.BtcScanner.ConfirmationsRequired,
		}); err != nil {
			log.WithError(err).Error()
		}
	}
}

// ExchangeStatusResponse http response for /api/exchange-status
type ExchangeStatusResponse struct {
	Error   string                        `json:"error"`
	Balance ExchangeStatusResponseBalance `json:"balance"`
}

// ExchangeStatusResponseBalance is the balance field of ExchangeStatusResponse
type ExchangeStatusResponseBalance struct {
	Kitties int `json:"kitties"`
}

// ExchangeStatusHandler returns the status of the exchanger
// Method: GET
// URI: /api/exchange-status
func ExchangeStatusHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		if !validMethod(ctx, w, r, []string{http.MethodGet}) {
			return
		}

		errorMsg := ""
		err := s.exchanger.Status()

		// If the status is an APIError, the most likely cause is that the
		// wallet has an insufficient balance (other causes could be a temporary
		// application error, or a bug in the skycoin node).
		// Errors that are not RPCErrors are transient and common, such as
		// exchange.ErrNotConfirmed, which will happen frequently and temporarily.
		switch err.(type) {
		case sender.RPCError:
			errorMsg = err.Error()
		default:
		}

		// Get the wallet balance,
		kittyBal, err := s.exchanger.Balance()
		kitties := 0
		if err != nil {
			log.WithError(err).Error("s.exchange.Balance failed")
		} else {
			kitties = kittyBal
		}

		resp := ExchangeStatusResponse{
			Error: errorMsg,
			Balance: ExchangeStatusResponseBalance{
				Kitties: kitties,
			},
		}

		log.WithField("resp", resp).Info()

		if err := httputil.JSONResponse(w, resp); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

// ReserveResponse represents the response of a reservation request
type ReserveResponse struct {
	DepositAddress string      `json:"deposit_address"`
	CoinType       string      `json:"coin_type"`
	Deadline       int64       `json:"deadline"`
	KittyID        iko.KittyID `json:"kitty_id"`
}

type reservationRequest struct {
	UserAddress      string `json:"user_address"`
	KittyID          string `json:"kitty_id"`
	CoinType         string `json:"coin_type"`
	VerificationCode string `json:"verification_code"`
}

// MakeReservationHandler handles kitty box reservations
// Method: POST
// Accept: application/json
// URI: /api/reservation/reserve
// Args:
//    {"user_address": "<user_address>", "kitty_id": "<kitty_id>", "coin_type": "<coin_type>", "verification_code": "<verification_code>"}
func MakeReservationHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// reserve can only be a post request
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			httputil.ErrResponse(w, http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			errorResponse(ctx, w, http.StatusUnsupportedMediaType, errors.New("Invalid content type"))
			return
		}

		reserveReq := &reservationRequest{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&reserveReq); err != nil {
			err = fmt.Errorf("Invalid json request body: %v", err)
			errorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}
		defer func(log logrus.FieldLogger) {
			if err := r.Body.Close(); err != nil {
				log.WithError(err).Warn("failed to closed request body")
			}
		}(log)

		// check that required parameters are given
		if reserveReq.UserAddress == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("missing user address"))
			return
		}

		if reserveReq.KittyID == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("missing kitty id"))
			return
		}

		if reserveReq.CoinType == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("missing cointype"))
			return
		}

		if reserveReq.VerificationCode == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("missing verification code"))
			return
		}

		_, err := cipher.DecodeBase58Address(reserveReq.UserAddress)
		if err != nil {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("invalid user address"))
			return
		}

		// Start a writable transaction.
		tx, err := s.db.Begin(true)
		if err != nil {
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}

		defer func() {
			if tx.DB() != nil {
				tx.Rollback()
			}
		}()

		log.Info("Calling service.BindAddress")
		boundAddr, err := s.service.BindAddressTx(tx, reserveReq.KittyID, reserveReq.CoinType)
		if err != nil {
			log.WithError(err).Error("service.BindAddress failed")
			switch err {
			case ErrBindDisabled:
				errorResponse(ctx, w, http.StatusForbidden, err)
			default:
				switch err {
				case addrs.ErrDepositAddressEmpty, ErrBoxAlreadyBound:
				default:
					err = errInternalServerError
				}
				errorResponse(ctx, w, http.StatusInternalServerError, err)
			}
			return
		}

		log.Info("Calling agent.MakeReservation")
		err = s.service.agentManager.MakeReservation(tx, boundAddr.Address, reserveReq.UserAddress,
			reserveReq.KittyID, reserveReq.CoinType, reserveReq.VerificationCode)
		if err != nil {
			log.WithError(err).Error("s.agent.MakeReservation failed")
			switch err {
			case agent.ErrMaxReservationsExceeded, agent.ErrBoxAlreadyReserved, agent.ErrInvalidCoinType:
				errorResponse(ctx, w, http.StatusBadRequest, err)
			default:
				errorResponse(ctx, w, http.StatusInternalServerError, err)
			}
			return
		}



		// get user from usermanager
		u, err := s.service.agentManager.UserManager.GetUser(reserveReq.UserAddress)
		if err != nil {
			log.WithError(err).Error("UserManager.GetUser failed")
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}
		// get the reservation for the reservation map
		reservation, err := s.service.agentManager.ReservationManager.GetReservationByKittyID(reserveReq.KittyID)
		if err != nil {
			log.WithError(err).Error("ReservationManager.GetReservationByKittyID failed")
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}

		ikoKittyID, err := iko.KittyIDFromString(reservation.KittyID)
		if err != nil {
			log.WithError(err).Error("iko.KittyIDFromString failed")
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}

		// update kitty api
		_, err = s.service.agentManager.KittyAPI.SetReservation(&rpc.ReservationIn{
			KittyID:     ikoKittyID,
			Reservation: reservation.Status,
		})
		if err != nil {
			log.WithError(err).Error("KittyAPI.SetReservation failed")
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}

		// satisfy the verification code
		err = s.service.agentManager.Verifier.SatisfyCode(reserveReq.VerificationCode, reservation.KittyID)
		if err != nil {
			log.WithError(err).Error("Verifier.SatisfyCode failed")
			httputil.ErrResponse(w, http.StatusInternalServerError)
			return
		}

		// add reservation to user
		err = s.service.agentManager.UserManager.AddReservation(u, reservation)
		if err != nil {
			log.WithError(err).Error("UserManager.AddReservation failed")
			return
		}

		// commit the transaction
		log.Info("commit reservation")
		tx.Commit()

		log = log.WithField("boundAddr", boundAddr)
		log.Infof("Bound sky and %s addresses", reserveReq.CoinType)

		kittyID, err := iko.KittyIDFromString(reserveReq.KittyID)
		if err != nil {
			errorResponse(ctx, w, http.StatusInternalServerError, err)
			log.WithError(err).Error()
			return
		}
		if err := httputil.JSONResponse(w, ReserveResponse{
			DepositAddress: boundAddr.Address,
			CoinType:       boundAddr.CoinType,
			Deadline:       time.Now().Add(time.Hour * 24).UnixNano(),
			KittyID:        kittyID,
		}); err != nil {
			errorResponse(ctx, w, http.StatusInternalServerError, err)
			log.WithError(err).Error()
			return
		}
	}
}

// ReservationsResponse represents reservations of a desired status
// like available/reserved/all
type ReservationsResponse struct {
	Reservations []agent.Reservation `json:"reservations"`
	Status       string              `json:"status"`
}

// GetReservationsHandler gets reservations based on the status
// Method: GET
// Accept: application/json
// URI: /api/reservation/getreservation?status=
// Args:
//    status: Reservation status, available/reserved/all
func GetReservationsHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// get reservations can only be a get request
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			httputil.ErrResponse(w, http.StatusMethodNotAllowed)
			return
		}

		status := r.URL.Query().Get("status")
		if status == "" {
			httputil.ErrResponse(w, http.StatusBadRequest)
			return
		}

		reservations, err := s.service.agentManager.GetReservations(status)
		if err != nil {
			log.WithError(err).Error("s.agent.GetReservations failed")
			errorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}

		if err := httputil.JSONResponse(w, ReservationsResponse{
			Reservations: reservations,
			Status:       status,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

// GetDepositAddressResponse represents response of get deposit address request
type GetDepositAddressResponse struct {
	DepositAddress string `json:"deposit_address"`
	KittyID        iko.KittyID `json:"kitty_id"`
}

// GetDepositAddressHandler gets deposit address of given kittyID
// Method: GET
// Accept: application/json
// URI: /api/reservation/getdepositaddress?kittyid=?
// Args:
//    kittyID: kitty ID of required kitty box
func GetDepositAddressHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		// get reservations can only be a get request
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			httputil.ErrResponse(w, http.StatusMethodNotAllowed)
			return
		}

		kittyID := r.URL.Query().Get("kittyid")
		if kittyID == "" {
			httputil.ErrResponse(w, http.StatusBadRequest)
			return
		}

		addr, err := s.service.agentManager.GetKittyDepositAddress(kittyID)
		if err != nil {
			log.WithError(err).Error("s.agent.GetKittyDepositAddress failed")
			errorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}

		kittyid, _ := iko.KittyIDFromString(kittyID)
		if err := httputil.JSONResponse(w, GetDepositAddressResponse{
			DepositAddress: addr,
			KittyID: kittyid,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}
