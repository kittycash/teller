package teller

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"math/big"
	"net/http"
	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/util/logger"
	"github.com/kittycash/teller/src/util/httputil"
	"github.com/kittycash/teller/src/sender"
	"encoding/json"
	"errors"
	"github.com/kittycash/teller/src/addrs"
	"github.com/kittycash/teller/src/agent"
	"github.com/skycoin/skycoin/src/util/droplet"
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
	Enabled                  bool   `json:"enabled"`
	BtcConfirmationsRequired int64  `json:"btc_confirmations_required"`
	EthConfirmationsRequired int64  `json:"eth_confirmations_required"`
	MaxBoundAddresses        int    `json:"max_bound_addrs"`
	SkyBtcExchangeRate       string `json:"sky_btc_exchange_rate"`
	SkyEthExchangeRate       string `json:"sky_eth_exchange_rate"`
	MaxDecimals              int    `json:"max_decimals"`
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

		// Convert the exchange rate to a skycoin balance string
		rate := s.cfg.BoxExchanger.BoxBtcExchangeRate
		maxDecimals := s.cfg.BoxExchanger.MaxDecimals
		dropletsPerBTC, err := exchange.CalculateBtcSkyValue(exchange.SatoshisPerBTC, rate, maxDecimals)
		if err != nil {
			log.WithError(err).Error("exchange.CalculateBtcSkyValue failed")
			errorResponse(ctx, w, http.StatusInternalServerError, errInternalServerError)
			return
		}

		skyPerBTC, err := droplet.ToString(dropletsPerBTC)
		if err != nil {
			log.WithError(err).Error("droplet.ToString failed")
			errorResponse(ctx, w, http.StatusInternalServerError, errInternalServerError)
			return
		}
		rate = s.cfg.BoxExchanger.BoxSkyExchangeRate
		dropletsPerETH, err := exchange.CalculateEthSkyValue(big.NewInt(exchange.WeiPerETH), rate, maxDecimals)
		if err != nil {
			log.WithError(err).Error("exchange.CalculateEthSkyValue failed")
			errorResponse(ctx, w, http.StatusInternalServerError, errInternalServerError)
			return
		}
		skyPerETH, err := droplet.ToString(dropletsPerETH)
		if err != nil {
			log.WithError(err).Error("droplet.ToString failed")
			errorResponse(ctx, w, http.StatusInternalServerError, errInternalServerError)
			return
		}

		if err := httputil.JSONResponse(w, ConfigResponse{
			Enabled:                  s.cfg.Teller.BindEnabled,
			BtcConfirmationsRequired: s.cfg.BtcScanner.ConfirmationsRequired,
			EthConfirmationsRequired: s.cfg.EthScanner.ConfirmationsRequired,
			SkyBtcExchangeRate:       skyPerBTC,
			SkyEthExchangeRate:       skyPerETH,
			MaxDecimals:              maxDecimals,
			MaxBoundAddresses:        s.cfg.Teller.MaxBoundAddresses,
		}); err != nil {
			log.WithError(err).Error(err)
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
	Coins string `json:"coins"`
	Hours string `json:"hours"`
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
		case sender.RPCError:
			errorMsg = err.Error()
		default:
		}

		// Get the wallet balance, but ignore any error. If an error occurs,
		// return a balance of 0
		bal, err := s.exchanger.Balance()
		coins := "0.000000"
		hours := "0"
		if err != nil {
			log.WithError(err).Error("s.exchange.Balance failed")
		} else {
			coins = bal.Coins
			hours = bal.Hours
		}

		resp := ExchangeStatusResponse{
			Error: errorMsg,
			Balance: ExchangeStatusResponseBalance{
				Coins: coins,
				Hours: hours,
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
		user := r.FormValue("user")
		if user == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing user address"))
			return
		}
		kitty := r.FormValue("kittyId")
		if kitty == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing kitty id"))
			return
		}
		cointype := r.FormValue("cointype")
		if cointype == "" {
			errorResponse(ctx, w, http.StatusBadRequest, errors.New("Missing cointype"))
			return
		}

		err := s.service.agentManager.DoReservation(reserveReq.UserAddress, reserveReq.KittyID, reserveReq.CoinType)
		if err != nil {
			log.WithError(err).Error("s.agent.DoReservation failed")
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
	KittyID string `json:"kitty_id"`
}

type cancelReservationRequest struct {
	KittyID string `json:"kitty_id"`
}

// cancelHandler cancels a reservation
// Method: POST
// Accept: application/json
// URI: /api/reservation/cancel
// Args:
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

		// kittyID will always be our unique identifier
		kittyID := r.FormValue("kittyId")
		if kittyID == "" {
			httputil.ErrResponse(w, http.StatusBadRequest)
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
		err := s.service.agentManager.CancelReservation(kittyID)
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
			KittyID: kittyID,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}

// ReservationsResponse represents reservations of a desired status, like available or reserved
type ReservationsResponse struct {
	Reservations []agent.Reservation `json:"reservations"`
	Status       string              `json:"status"`
}

// GetReservationsHandler gets reservations based on the status
// Method: GET
// Accept: application/json
// URI: /api/reservation/getreservation?status=
// Args:
//    status
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

		if err := httputil.JSONResponse(w, reservations); err != nil {
			log.WithError(err).Error(err)
		}

		if err := httputil.JSONResponse(w, ReservationsResponse{
			Reservations: reservations,
			Status: status,
		}); err != nil {
			log.WithError(err).Error(err)
		}
	}
}