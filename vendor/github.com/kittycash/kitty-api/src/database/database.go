package database

import (
	"encoding/json"
	"fmt"

	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"github.com/skycoin/skycoin/src/cipher"
	"golang.org/x/net/context"
)

type DBPublic interface {
	Count(ctx context.Context) (int64, error)

	GetEntryOfID(ctx context.Context, kittyID iko.KittyID) (*Entry, error)
	GetEntryOfDNA(ctx context.Context, kittyDNA string) (*Entry, error)

	GetEntries(ctx context.Context,
		startIndex, pageSize int,
		filters *Filters, sorters *Sorters) (int64, []*Entry, error)
}

type Database interface {
	Add(ctx context.Context, entry *Entry) error
	MultiAdd(ctx context.Context, entries []*Entry) error

	Count(ctx context.Context) (int64, error)

	GetEntryOfID(ctx context.Context, kittyID iko.KittyID) (*Entry, error)
	GetEntryOfDNA(ctx context.Context, kittyDNA string) (*Entry, error)

	GetEntries(ctx context.Context,
		startIndex, pageSize int,
		filters *Filters, sorters *Sorters) (int64, []*Entry, error)

	SetReservationOfEntry(ctx context.Context,
		kittyID iko.KittyID, reservation string) (*Entry, error)
}

// Filter is used to filter a set of entries.
type Filter struct {
	Unit string // What unit are we filtering in?
	Min  int64  // default = 0
	Max  int64  // default = 9223372036854775807
}

func (pf *Filter) Check() error {
	if pf.Min > pf.Max {
		return errors.New("filter minimum value cannot be greater than it's maximum value")
	}
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

func (f *Filters) Len() int {
	return len(f.m)
}

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

type FilterAction func(key string, filter Filter) error

func (f *Filters) Range(action FilterAction) error {
	for k, filter := range f.m {
		if err := action(k, filter); err != nil {
			return err
		}
	}
	return nil
}

func (f *Filters) GetKeys() []string {
	keys := make([]string, len(f.m))
	var i int
	for k := range f.m {
		keys[i], i = k, i+1
	}
	return keys
}

type Sorter string

type SortDirection byte

var (
	SortAsc  = SortDirection(0)
	SortDesc = SortDirection(1)
)

func (s Sorter) Extract() (SortDirection, string) {
	switch len(s) {
	case 0:
		return SortDesc, ""
	default:
		switch s[0] {
		case '+':
			return SortDesc, string(s[1:])
		case '-':
			return SortAsc, string(s[1:])
		default:
			return SortDesc, string(s)
		}
	}
}

type Sorters struct {
	a []Sorter
	m map[Sorter]struct{}
}

func NewSorters() *Sorters {
	return &Sorters{
		m: make(map[Sorter]struct{}),
	}
}

func (s *Sorters) Len() int {
	return len(s.a)
}

func (s *Sorters) Add(v Sorter) error {
	if _, ok := s.m[v]; ok {
		return errors.Errorf("sorter for '%s' is redefined", v)
	}
	s.m[v] = struct{}{}
	s.a = append(s.a, v)
	return nil
}

type SorterAction func(index int, sorter Sorter) error

func (s *Sorters) Range(action SorterAction) error {
	for i, sorter := range s.a {
		if err := action(i, sorter); err != nil {
			return errors.WithMessage(err,
				fmt.Sprintf("failed on index '%d'", i))
		}
	}
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

	PriceBTC    int64  `json:"price_btc"`   // Price of kitty in BTC.
	PriceSKY    int64  `json:"price_sky"`   // Price of kitty in SKY.

	BoxOpen   bool   `json:"box_open"`    // Whether box is open.
	BirthDate int64  `json:"birth_date"` // Timestamp of box opening.
	KittyDNA  string `json:"kitty_dna"`  // Hex representation of kitty DNA (after box opening).

	BoxImgURL   string `json:"box_image_url"`   // Box image URL.
	KittyImgURL string `json:"kitty_image_url"` // Kitty image URL.
}
*/

type Entry struct {
	iko.Kitty
	Sig         string `json:"sig"`
	Reservation string `json:"reservation"` // Whether kitty is reserved or not.
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

func (e *Entry) Sign(sk cipher.SecKey) {
	e.Sig = e.Kitty.Sign(sk).Hex()
}

func (e *Entry) Verify(pk cipher.PubKey) error {
	sig, err := cipher.SigFromHex(e.Sig)
	if err != nil {
		return err
	}
	return e.Kitty.Verify(pk, sig)
}
