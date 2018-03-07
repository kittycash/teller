package teller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kittycash/teller/src/addrs"
	"github.com/kittycash/teller/src/agent"
	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/teller/src/util/httputil"
	"github.com/kittycash/teller/src/util/logger"
	"github.com/sirupsen/logrus"
)

//@TODO add authentication checks

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
		//@TODO (therealssj): implement
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

		// If the status is an RPCError, the most likely cause is that the
		// wallet has an insufficient balance (other causes could be a temporary
		// application error, or a bug in the skycoin node).
		// Errors that are not RPCErrors are transient and common, such as
		// exchange.ErrNotConfirmed, which will happen frequently and temporarily.
		switch err.(type) {
		case sender.APIError:
			errorMsg = err.Error()
		default:
		}

		// Get the wallet balance, 
		kitties := s.exchanger.Balance()

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
	DepositAddress string `json:"deposit_address"`
	PaymentAmount  string `json:"payment_amount"`
	CoinType       string `json:"coin_type"`
}

type reservationRequest struct {
	UserAddress string `json:"user_address"`
	KittyID     string `json:"kitty_id"`
	CoinType    string `json:"coin_type"`
}

// ReserveHandler handles kitty box reservations
// Method: POST
// Accept: application/json
// URI: /api/reservation/reserve
// Args:
//    {"user_address": "<user_address>", "kitty_id": "<kitty_id>", "coin_type": "<coin_type>"}
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
				log.WithError(err).Warn("Failed to closed request body")
			}
		}(log)

		// check that required parameters are given
		if reserveReq.UserAddress == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing user address"))
			return
		}

		if reserveReq.KittyID == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing kitty id"))
			return
		}

		if reserveReq.CoinType == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing cointype"))
			return
		}

		err := s.service.agentManager.MakeReservation(reserveReq.UserAddress, reserveReq.KittyID, reserveReq.CoinType)
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

		log.Info("Calling service.BindAddress")

		boundAddr, err := s.service.BindAddress(reserveReq.KittyID, reserveReq.CoinType)
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

		log = log.WithField("boundAddr", boundAddr)
		log.Infof("Bound sky and %s addresses", reserveReq.CoinType)

		if err := httputil.JSONResponse(w, ReserveResponse{
			DepositAddress: boundAddr.Address,
			CoinType:       boundAddr.CoinType,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

//@TODO do I need this? return more information?
// CancelReservationResponse represents response of a cancel reservation request
type CancelReservationResponse struct {
	UserAddress string `json:"user_address"`
	KittyID     string `json:"kitty_id"`
}

type cancelReservationRequest struct {
	UserAddress string `json:"user_address"`
	KittyID     string `json:"kitty_id"`
}

// cancelHandler cancels a reservation
// Method: POST
// Accept: application/json
// URI: /api/reservation/cancel
// Args:
//    {"user_address": "<user_address>"}
//    {"kitty_id": "<kitty_Id>"}
func CancelReservationHandler(s *HTTPServer) http.HandlerFunc {
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

		cancelReservationReq := &cancelReservationRequest{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&cancelReservationReq); err != nil {
			err = fmt.Errorf("Invalid json request body: %v", err)
			errorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}
		defer func(log logrus.FieldLogger) {
			if err := r.Body.Close(); err != nil {
				log.WithError(err).Warn("Failed to closed request body")
			}
		}(log)

		// cancel the reservation
		err := s.service.agentManager.CancelReservation(
		cancelReservationReq.UserAddress, cancelReservationReq.KittyID)

		if err != nil {
			log.WithError(err).Error("s.agent.CancelReservation failed")
			switch err {
			case agent.ErrReservationNotFound:
				errorResponse(ctx, w, http.StatusBadRequest, err)
			default:
				errorResponse(ctx, w, http.StatusInternalServerError, err)
			}
			return
		}

		if err := httputil.JSONResponse(w, CancelReservationResponse{
			cancelReservationReq.UserAddress,
			cancelReservationReq.KittyID,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

// ReservationsResponse represents reservations of a desired status
// like available or reserved
type ReservationsResponse struct {
	Reservations []agent.Reservation `json:"reservations"`
	Status       string              `json:"status"`
}

// GetReservationsHandler gets reservations based on the status
// Method: GET
// Accept: application/json
// URI: /api/reservation/getreservation?status=
// Args:
//    status: Reservation status, available or reserved
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
		}

		if err := httputil.JSONResponse(w, ReservationsResponse{
			Reservations: reservations,
			Status:       status,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

type GetDepositAddressResponse struct {
	UserAddress string `json:"user_address"`
	KittyID     string `json:"kitty_id"`
}

// GetReservationsHandler gets reservations based on the status
// Method: GET
// Accept: application/json
// URI: /api/reservation/getdepositaddress?useraddr=?&kittyid=?
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

		userAddr := r.URL.Query().Get("useraddr")
		if userAddr == "" {
			httputil.ErrResponse(w, http.StatusBadRequest)
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
		}

		if err := httputil.JSONResponse(w, GetDepositAddressResponse{
			UserAddress: addr,
			KittyID:     kittyID,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}
