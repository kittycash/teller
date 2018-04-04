package http

import (
	"github.com/kittycash/wallet/src/kitties"
	"net/http"
	"fmt"
	"github.com/kittycash/wallet/src/iko"
	"io/ioutil"
)

func marketKitties(m *http.ServeMux, g *kitties.Manager, bc *iko.BlockChain) error {
	Handle(m, "/api/count", http.MethodGet, count(g))
	Handle(m, "/api/entry", http.MethodGet, entry(g, bc))
	Handle(m, "/api/entries", http.MethodGet, entries(g, bc))
	return nil
}

func count(g *kitties.Manager) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Count(req)
	})
}

func entry(g *kitties.Manager, bc *iko.BlockChain) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Entry(bc, req)
	})
}

func entries(g *kitties.Manager, bc *iko.BlockChain) HandlerFunc {
	return marketHandler(func(req *http.Request) (*http.Response, error) {
		return g.Entries(bc, req)
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type MHAction func(req *http.Request) (*http.Response, error)

func marketHandler(action MHAction) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		resp, err := action(r)
		if err != nil {
			return sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		data, _ := ioutil.ReadAll(resp.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, err = w.Write(data)
		return err
	}
}