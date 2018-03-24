package api

import (
	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"strings"
)

/*
	<<< ENDPOINT: count >>>
*/

type CountOut struct {
	Count int64 `json:"count"`
}

/*
	<<< ENDPOINT: entry >>>
*/

type EntryIn struct {
	UseKittyID  bool
	KittyID     iko.KittyID
	UseKittyDNA bool
	KittyDNA    string
}

func (e *EntryIn) Check() error {
	if e.UseKittyID == e.UseKittyDNA {
		return errors.New("no kitty_id or kitty_dna is provided")
	}
	if e.UseKittyDNA {
		// TODO: Check KittyDNA.
	}
	return nil
}

func GetEntryIn(qs url.Values) (*EntryIn, error) {
	var (
		err error
		in  = new(EntryIn)

		kittyID  = qs.Get("kitty_id")
		kittyDNA = qs.Get("kitty_dna")
	)
	if kittyID != "" {
		in.UseKittyID = true
		id, err := strconv.ParseUint(kittyID, 10, 64)
		if err != nil {
			return nil, err
		}
		in.KittyID = iko.KittyID(id)
	}
	if kittyDNA != "" {
		in.UseKittyDNA = true
		in.KittyDNA = kittyDNA
	}
	return in, err
}

type EntryOut struct {
	Entry *database.Entry
}

/*
	<<< ENDPOINT: entries >>>
*/

var (
	filterPriceUnits = map[string]struct{}{
		"btc": {},
		"sky": {},
	}
)

type EntriesIn struct {
	Filters    *database.Filters // nil = no filters
	Order      *database.Sorters // nil = default order (by kittyID)
	StartIndex int               `default:"0"`
	PageSize   int               `default:"10"`
}

func (e *EntriesIn) checkPage() error {
	switch {
	case e.PageSize < 1:
		return errors.New("")
	}
	if e.PageSize < 1 {

	}
	// TODO: Implement.
	return nil
}

func GetEntriesIn(qs url.Values) (*EntriesIn, error) {
	var (
		err         error
		in          = new(EntriesIn)
		filterPrice = qs.Get("filter_price")
		filterDate  = qs.Get("filter_date")
		order       = qs.Get("order")
		startIndex  = qs.Get("start_index")
		pageSize    = qs.Get("page_size")
	)

	if hasPF, hasDF := filterPrice != "", filterDate != ""; hasPF || hasDF {
		in.Filters = database.NewFilters()

		if hasPF {
			var parts = strings.Split(filterPrice, ",")
			if len(parts) != 3 {
				return nil, errors.Errorf(
					"invalid '%s' query, expected three elements of format '%s'",
					"filter_price", "{currency_unit},{min_value},{max_value}")
			}
			min, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err,
					"invalid '%s' query, expected minimum element [0] to be numerical",
					"filter_price")
			}
			max, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err,
					"invalid '%s' query, expected maximum element [1] to be numerical",
					"filter_price")
			}
			var unit = strings.ToLower(parts[2])
			if _, ok := filterPriceUnits[unit]; !ok {
				return nil, errors.Errorf(
					"invalid '%s' query, expected unit element [2] to be %s",
					"filter_price", "either 'btc' or 'sky'")
			}
			filter := database.Filter{
				Min:  min,
				Max:  max,
				Unit: unit,
			}
			if err = in.Filters.Add("price", filter); err != nil {
				return nil, err
			}
		}

		if hasDF {
			var parts = strings.Split(filterDate, ",")
			min, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err,
					"invalid '%s' query, expected minimum element [0] to be numerical",
					"filter_date")
			}
			max, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err,
					"invalid '%s' query, expected maximum element [1] to be numerical",
					"filter_date")
			}
			filter := database.Filter{
				Min: min,
				Max: max,
			}
			if err := in.Filters.Add("date", filter); err != nil {
				return nil, err
			}
		}
	}

	if order != "" {
		in.Order = database.NewSorters()
		for _, v := range strings.Split(order, ",") {
			if err := in.Order.Add(database.Sorter(v)); err != nil {
				return nil, err
			}
		}
	}

	in.StartIndex, err = strconv.Atoi(startIndex)
	if err != nil {
		return nil, err
	}

	in.PageSize, err = strconv.Atoi(pageSize)
	if err != nil {
		return nil, err
	}

	return in, nil
}

type EntriesOut struct {
	Count   int               `json:"count"`
	Entries []*database.Entry `json:"entries"`
}
