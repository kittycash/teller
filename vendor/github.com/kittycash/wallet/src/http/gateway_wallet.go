package http

import (
	"net/http"
	"github.com/kittycash/wallet/src/wallet"
)

func walletGateway(mux *http.ServeMux, g *wallet.Manager) error {

	Handle(mux, "/api/wallets/refresh",
		"GET", refreshWallets(g))

	Handle(mux, "/api/wallets/list",
		"GET", listWallets(g))

	Handle(mux, "/api/wallets/new",
		"POST", newWallet(g))

	return nil
}

func refreshWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		// TODO: implement.
		// CALLS: 'g.Refresh()'.
		// RESPONSE:
		// - 'Content-Type': 'application/json'.
		// - true & 200   : on success.
		// - string & 500 : of the error on failure.
		return nil
	}
}

type WalletsReply struct {
	Wallets []wallet.Stat `json:"wallets"`
}

func listWallets(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		// TODO: implement.
		// CALLS: 'g.ListWallets()'.
		// RESPONSE:
		// - 'Content-Type': 'application/json'.
		// - json representation of 'WalletsReply'.
		return nil
	}
}

func newWallet(g *wallet.Manager) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		// TODO: implement.
		// CALLS: 'g.NewWallet()'.
		// REQUEST:
		// - 'Content-Type': 'application/x-www-form-urlencoded'.
		// - Check actual function for all key-values.
		// RESPONSE:
		// - 'Content-Type': 'application/json'.
		// - true & 200   : on success.
		// - string & http status code : of the error on failure.
		return nil
	}
}

