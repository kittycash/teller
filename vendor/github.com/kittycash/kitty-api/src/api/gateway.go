package api

import (
	"fmt"
	"net/http"

	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/wallet/src/iko"
)

func ServePublic(db database.DBPublic) *http.ServeMux {
	var m = http.NewServeMux()
	Handle(m, "/api/count", "GET", count(db))
	Handle(m, "/api/entry", "GET", entry(db))
	Handle(m, "/api/entries", "GET", entries(db))
	return m
}

func count(db database.DBPublic) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		count, err := db.Count(r.Context())
		if err != nil {
			return sendPrivateJson(w, http.StatusInternalServerError,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		return sendPublicJson(w, http.StatusOK, CountOut{
			Count: count,
		})
	}
}

func entry(db database.DBPublic) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		in, err := GetEntryIn(r.URL.Query())
		if err != nil {
			return sendPrivateJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		if err := in.Check(); err != nil {
			return sendPrivateJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		var (
			entry *iko.KittyEntry
		)
		switch {
		case in.UseKittyID:
			entry, err = db.GetEntryOfID(r.Context(), in.KittyID)
		case in.UseKittyDNA:
			entry, err = db.GetEntryOfDNA(r.Context(), in.KittyDNA)
		}
		if err != nil {
			return sendPrivateJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		return sendPublicJson(w, http.StatusOK, EntryOut{
			Entry: entry,
		})
	}
}

func entries(db database.DBPublic) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		in, err := GetEntriesIn(r.URL.Query())
		if err != nil {
			return sendPrivateJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		total, entries, err := db.GetEntries(r.Context(),
			in.Offset, in.PageSize, in.Filters, in.Order)
		if err != nil {
			return sendPrivateJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", err.Error()))
		}
		return sendPublicJson(w, http.StatusOK, EntriesOut{
			TotalCount: total,
			PageCount:  len(entries),
			Entries:    entries,
		})
	}
}
