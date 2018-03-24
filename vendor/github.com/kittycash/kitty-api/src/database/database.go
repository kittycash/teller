package database

import (
	"encoding/json"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Database interface {
	Add(ctx context.Context, entry *Entry) error
	Count(ctx context.Context) (int64, error)

	GetEntryOfID(ctx context.Context, kittyID iko.KittyID) (*Entry, error)
	GetEntryOfDNA(ctx context.Context, kittyDNA string) (*Entry, error)

	GetEntries(ctx context.Context,
		startIndex, pageSize int,
		filters *Filters, sorters *Sorters) ([]*Entry, error)

	SetReservationOfEntry(ctx context.Context,
		kittyID iko.KittyID, isReserved bool) (*Entry, error)
}

// Filter is used to filter a set of entries
type Filter struct {
	Unit string // What unit are we filtering in?
	Min  int64  // default = 0
	Max  int64  // default = 9223372036854775807
}

func (pf *Filter) Check() error {
	// TODO: Implement.
	return nil
}

// Filters represents a list of filters.
type Filters struct {
	m map[string]Filter
}

// NewFilters creates a new filter.
func NewFilters() *Filters {
	return &Filters{
		m: make(map[string]Filter),
	}
}

var (
	filterKeys = map[string]struct{}{
		"price": {},
		"date":  {},
	}
)

func (f *Filters) Add(k string, v Filter) error {
	if _, ok := filterKeys[k]; !ok {
		return errors.Errorf("cannot filter for '%s'", k)
	}
	if err := v.Check(); err != nil {
		return errors.Wrapf(err, "filter for '%s' is invalid", k)
	}
	if _, ok := f.m[k]; ok {
		return errors.Errorf("filter for '%s' is re-defined", k)
	}
	f.m[k] = v
	return nil
}

type Sorter string

type Sorters struct {
	a []Sorter
	m map[Sorter]struct{}
}

func NewSorters() *Sorters {
	return &Sorters{
		m: make(map[Sorter]struct{}),
	}
}

func (s *Sorters) Add(v Sorter) error {
	if _, ok := s.m[v]; ok {
		return errors.Errorf("sorter for '%s' is redefined", v)
	}
	s.m[v] = struct{}{}
	return nil
}

/*
	<<< ENTRY >>>
*/

/*
type Kitty struct {
	ID    KittyID `json:"kitty_id"`    // Identifier for kitty.
	Name  string  `json:"name"`        // Name of kitty.
	Desc  string  `json:"description"` // Description of kitty.
	Breed string  `json:"breed"`       // Kitty breed.

	PriceBTC   int64 `json:"price_btc"`   // Price of kitty in BTC.
	PriceSKY   int64 `json:"price_sky"`   // Price of kitty in SKY.
	IsReserved bool  `json:"is_reserved"` // Whether kitty is reserved.

	IsOpen    bool   `json:"is_open"`    // Whether box is open.
	BirthDate int64  `json:"birth_date"` // Timestamp of box opening.
	KittyDNA  string `json:"kitty_dna"`  // Hex representation of kitty DNA (after box opening).

	BoxImgURL   string `json:"box_image_url"`   // Box image URL.
	KittyImgURL string `json:"kitty_image_url"` // Kitty image URL.
}
*/

type Entry struct {
	iko.Kitty
}

func EntryFromJson(raw []byte) (*Entry, error) {
	out := new(Entry)
	err := json.Unmarshal(raw, out)
	return out, err
}

func (e *Entry) Json() []byte {
	raw, _ := json.Marshal(e)
	return raw
}
