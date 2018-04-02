package rpc

import (
	"context"
	"fmt"

	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"github.com/skycoin/skycoin/src/cipher"
)

type Gateway struct {
	pk cipher.PubKey
	db database.Database
}

type EntryIn struct {
	Entry *iko.KittyEntry
}

func (g *Gateway) AddEntry(in *EntryIn, _ *struct{}) error {
	if err := checkEntry(in.Entry, g.pk); err != nil {
		return err
	}
	return g.db.Add(context.Background(), in.Entry)
}

type AddEntriesIn struct {
	Entries []*iko.KittyEntry
}

func (g *Gateway) AddEntries(in *AddEntriesIn, _ *struct{}) error {
	for i, entry := range in.Entries {
		if err := checkEntry(entry, g.pk); err != nil {
			return errors.WithMessage(err, fmt.Sprintf(
				"failed at index '%d'", i))
		}
	}
	return g.db.MultiAdd(context.Background(), in.Entries)
}

type CountOut struct {
	Count int64
}

func (g *Gateway) Count(_ *struct{}, out *CountOut) error {
	count, err := g.db.Count(context.Background())
	if err != nil {
		return err
	}
	out.Count = count
	return nil
}

type EntryOfIDIn struct {
	KittyID iko.KittyID
}

type EntryOfIDOut struct {
	Entry *iko.KittyEntry
}

func (g *Gateway) EntryOfID(in *EntryOfIDIn, out *EntryOfIDOut) error {
	entry, err := g.db.GetEntryOfID(context.Background(), in.KittyID)
	if err != nil {
		return err
	}
	out.Entry = entry
	return nil
}

type EntryOfDNAIn struct {
	KittyDNA string
}

type EntryOfDNAOut struct {
	Entry *iko.KittyEntry
}

func (g *Gateway) EntryOfDNA(in *EntryOfDNAIn, out *EntryOfDNAOut) error {
	entry, err := g.db.GetEntryOfDNA(context.Background(), in.KittyDNA)
	if err != nil {
		return err
	}
	out.Entry = entry
	return nil
}

type EntriesIn struct {
	Offset   int
	PageSize int
	Query    string
	Filters  database.RPCFilters
	Sorters  database.RPCSorters
}

type EntriesOut struct {
	TotalCount int64
	Results    []*iko.KittyEntry
}

func (g *Gateway) Entries(in *EntriesIn, out *EntriesOut) error {
	filters, err := in.Filters.ToFilters()
	if err != nil {
		return errors.WithMessage(err,
			"failed to generate filters")
	}
	sorters, err := in.Sorters.ToSorters()
	if err != nil {
		return errors.WithMessage(err,
			"failed to generate sorters")
	}
	count, res, err := g.db.GetEntries(context.Background(),
		in.Offset, in.PageSize, in.Query, filters, sorters)
	if err != nil {
		return err
	}
	out.TotalCount = count
	out.Results = res
	return nil
}

type ReservationIn struct {
	KittyID     iko.KittyID
	Reservation string
}

type ReservationOut struct {
	Entry *iko.KittyEntry
}

func (g *Gateway) SetReservation(in *ReservationIn, out *ReservationOut) error {
	entry, err := g.db.SetReservationOfEntry(
		context.Background(), in.KittyID, in.Reservation)
	if err != nil {
		return err
	}
	out.Entry = entry
	return nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func checkEntry(entry *iko.KittyEntry, pk cipher.PubKey) error {
	if err := entry.CheckData(); err != nil {
		return err
	}
	if err := entry.Verify(pk); err != nil {
		return err
	}
	return nil
}