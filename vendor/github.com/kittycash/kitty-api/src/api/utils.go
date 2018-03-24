package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handle(mux *http.ServeMux, pattern, method string, handler HandlerFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("invalid method type of '%s', expected '%s'",
					r.Method, method))
		} else if err := handler(w, r); err != nil {
			fmt.Println(err)
		}
	})
}

func sendJson(w http.ResponseWriter, status int, v interface{}) error {
	data, e := json.Marshal(v)
	if e != nil {
		return e
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, e = w.Write(data)
	return e
}
