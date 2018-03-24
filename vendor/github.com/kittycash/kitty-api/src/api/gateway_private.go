package api

import (
	"encoding/json"
	"fmt"
	"github.com/kittycash/kitty-api/src/database"
	"io/ioutil"
	"net/http"
)

func ServePrivate(db database.Database) *http.ServeMux {
	var m = http.NewServeMux()
	Handle(m, "/api/add", "POST", add(db))
	return m
}

func add(db database.Database) HandlerFunc {

	// TODO: Implement way to restrict access.

	return func(w http.ResponseWriter, r *http.Request) error {
		switch ct := r.Header.Get("Content-Type"); ct {
		case "application/json":
			raw, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return sendJson(w, http.StatusBadRequest,
					fmt.Sprintf("Error: %s", err.Error()))
			}
			entry := new(database.Entry)
			if err := json.Unmarshal(raw, entry); err != nil {
				return sendJson(w, http.StatusBadRequest,
					fmt.Sprintf("Error: %s", err.Error()))
			}
			if err := db.Add(r.Context(), entry); err != nil {
				return sendJson(w, http.StatusBadRequest,
					fmt.Sprintf("Error: %s", err.Error()))
			}
			return sendJson(w, http.StatusOK,
				true)
		default:
			return sendJson(w, http.StatusBadRequest,
				"invalid request 'Content-Type'")
		}
	}
}
