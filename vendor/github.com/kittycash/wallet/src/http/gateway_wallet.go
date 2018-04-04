package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kittycash/wallet/src/wallet"
)

func walletGateway(m *http.ServeMux, g *wallet.Manager) error {
	Handle(m, "/api/wallets/refresh", "GET", refreshWallets(g))
	Handle(m, "/api/wallets/list", "GET", listWallets(g))
	Handle(m, "/api/wallets/new", "POST", newWallet(g))
	Handle(m, "/api/wallets/delete", "POST", deleteWallet(g))
	Handle(m, "/api/wallets/get", "POST", getWallet(g))
	Handle(m, "/api/wallets/seed", "POST", newSeed())
	return nil
}

func refreshWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		// Send json response with 500 status code if error.
		if e := g.Refresh(); e != nil {
			sendJson(w, http.StatusInternalServerError,
				fmt.Sprintf("message: '%s'", e))
		}
		// Send json response with 200 status code if error is nil.
		return sendJson(w, http.StatusOK, true)
	}
}

type WalletsReply struct {
	Wallets []wallet.Stat `json:"wallets"`
}

func listWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		return sendJson(w, http.StatusOK, WalletsReply{
			Wallets: g.ListWallets(),
		})
	}
}

func newWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel     = r.PostFormValue("label")
					vSeed      = r.PostFormValue("seed")
					vAddresses = r.PostFormValue("aCount")
					vEncrypted = r.PostFormValue("encrypted")
					vPassword  = r.PostFormValue("password")
				)

				encrypted, e := strconv.ParseBool(vEncrypted)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e))
				}

				// Options to pass to g.NewWallet()
				opts := wallet.Options{
					Label:     vLabel,
					Seed:      vSeed,
					Encrypted: encrypted,
					Password:  vPassword,
				}

				/**
				 * Verify that all values are correct
				 * Respond if options are not correct
				 */

				if e := opts.Verify(); e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				// Get aCount and convert it to int.
				aCount, e := strconv.Atoi(vAddresses)
				if e != nil {
					return false, sendJson(w, http.StatusNotAcceptable,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				if e := g.NewWallet(&opts, aCount); e != nil {
					return false, sendJson(w, http.StatusInternalServerError,
						fmt.Sprintf("Error: %s", e.Error()))
				}

				return true, sendJson(w, http.StatusOK, true)
			},
		})
		return e
	}
}

func deleteWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel = r.PostFormValue("label")
				)
				if r.Body == nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprint("request body missing"))
				}
				if e := g.DeleteWallet(vLabel); e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: failed to delete wallet of label '%s': %v",
							vLabel, e))
				}
				return true, sendJson(w, http.StatusOK, true)
			},
		})
		return e
	}
}

func getWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				var (
					vLabel     = r.PostFormValue("label")
					vPassword  = r.PostFormValue("password") // Optional.
					vAddresses = r.PostFormValue("aCount")   // Optional.
				)
				if r.Body == nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprint("request body missing"))
				}
				var addresses int
				if vAddresses != "" {
					var e error
					addresses, e = strconv.Atoi(vAddresses)
					if e != nil {
						return false, sendJson(w, http.StatusBadRequest,
							fmt.Sprintf("Error: %s", e))
					}
				}
				fw, e := g.DisplayWallet(vLabel, vPassword, addresses)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %v", e))
				}
				return true, sendJson(w, http.StatusOK, fw)
			},
		})
		return e
	}
}

type SeedReply struct {
	Seed string `json:"seed"`
}

func newSeed() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {

		// Only allow 'Content-Type' of 'application/x-www-form-urlencoded'.
		_, e := SwitchContType(w, r, ContTypeActions{
			CtApplicationForm: func() (bool, error) {
				vSeedBitSize := r.PostFormValue("seedBitSize")
				seedBitSize, e := wallet.SeedBitSizeFromString(vSeedBitSize)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						fmt.Sprintf("Error: %s", e))
				}
				seed, e := wallet.NewSeed(seedBitSize)
				if e != nil {
					return false, sendJson(w, http.StatusInternalServerError,
						fmt.Sprintf("Error: %v", e))
				}
				return true, sendJson(w, http.StatusOK, SeedReply{
					Seed: seed,
				})
			},
		})
		return e
	}
}
