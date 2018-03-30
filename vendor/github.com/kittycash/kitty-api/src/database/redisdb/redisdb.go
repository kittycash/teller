package redisdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
)

const (
	EntryPrefix    = "entry"
	EntryIndexName = "entryIndex"
)

func EntryKey(id iko.KittyID) string {
	return fmt.Sprintf("%s:%d", EntryPrefix, id)
}

type Config struct {
	Address  string
	TestMode bool
	Database int
	Password string
}

type RedisDB struct {
	conn redis.Conn
	mux  sync.Mutex
}

func New(config *Config) (*RedisDB, error) {

	options := []redis.DialOption{
		redis.DialDatabase(config.Database),
		redis.DialConnectTimeout(time.Second * 10),
	}
	if config.Password != "" {
		options = append(options,
			redis.DialPassword(config.Password))
	}
	conn, err := redis.Dial("tcp", config.Address, options...)
	if err != nil {
		return nil, errors.Wrap(err, "dial to redis server failed")
	}
	if config.TestMode {
		fmt.Println("WARNING: running in test mode!")
		if err := dropConn(conn); err != nil {
			fmt.Println("DROP OLD DATA: (error)", err)
		} else {
			fmt.Println("DROP OLD DATA: (success)")
		}
	}
	if err := initConn(conn); err != nil {
		return nil, errors.Wrap(err, "failed to init with 'FT.CREATE'")
	}
	return &RedisDB{
		conn: conn,
	}, nil
}

func dropConn(conn redis.Conn) error {
	_, err := conn.Do("FT.DROP", EntryIndexName, 0)
	return err
}

func initConn(conn redis.Conn) error {
	if _, err := conn.Do("FT.INFO", EntryIndexName); err != nil {
		// Index not created, attempt to create index.
		_, err = conn.Do(
			"FT.CREATE",
			EntryIndexName, "NOHL", "SCHEMA",

			"kitty_id", "NUMERIC", "SORTABLE",
			"name", "TEXT", "SORTABLE", "NOINDEX",
			"dna", "TEXT", "NOSTEM",

			"price_btc", "NUMERIC", "SORTABLE",
			"price_sky", "NUMERIC", "SORTABLE",
			"reservation", "TEXT", "SORTABLE", "NOINDEX",

			"created", "NUMERIC", "SORTABLE",
			"raw", "TEXT", "NOSTEM", "NOINDEX",
		)
		if err != nil {
			return errors.Wrap(err, "failed to create index")
		}
	}
	return nil
}

func (r *RedisDB) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RedisDB) Add(_ context.Context, entry *database.Entry) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.addEntry(EntryKey(entry.ID), entry.Json(), entry)
}

func (r *RedisDB) MultiAdd(ctx context.Context, entries []*database.Entry) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for i, entry := range entries {
		args := r.getAddEntryArgs(EntryKey(entry.ID), entry.Json(), entry)

		if out, err := r.conn.Do("FT.ADD", args...); err != nil {
			err = errors.WithMessage(err, "internal server error")
			fmt.Printf("(*RedisDB).MultiAdd [%d] Error: %s", i, err.Error())
		} else {
			outBytes, _ := json.MarshalIndent(out, "", "    ")
			fmt.Printf("(*RedisDB).MultiAdd RESULT[%d]: %s\n", i, string(outBytes))
		}
	}

	return nil
}

func (r *RedisDB) Count(ctx context.Context) (int64, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	const DocsCountIndex = 7

	out, err := r.conn.Do("FT.INFO", EntryIndexName)
	if err != nil {
		return -1, err
	}

	return redis.Int64(out.([]interface{})[DocsCountIndex], nil)
}

func (r *RedisDB) GetEntryOfID(_ context.Context, kittyID iko.KittyID) (*database.Entry, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	key := EntryKey(kittyID)

	entry, err := r.getRawEntry(key)
	if err != nil {
		return nil, err
	}

	return database.EntryFromJson(entry)
}

func (r *RedisDB) GetEntryOfDNA(_ context.Context, kittyDNA string) (*database.Entry, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	out, err := r.conn.Do(
		"FT.SEARCH",
		EntryIndexName, kittyDNA,
		"INFIELDS", 1, "dna",
		"LIMIT", 0, 1,
	)
	if err != nil {
		return nil, err
	}
	data, ok := out.([]interface{})
	if !ok {
		return nil, errors.New(
			"internal server error: results format is unexpected")
	}
	if dLen := len(data); dLen != 3 {
		return nil, errors.Errorf(
			"failed to fetch entry of given DNA '%s'", kittyDNA)
	}
	entry, ok := data[len(data)-1].([]interface{})
	if !ok {
		return nil, errors.New(
			"internal server error: conversion of 'result -> []interface{}'")
	}
	raw, ok := entry[len(entry)-1].([]byte)
	if !ok {
		return nil, errors.New(
			"internal server error: conversion of 'interface{} -> []byte'")
	}
	return database.EntryFromJson(raw)
}

func (r *RedisDB) GetEntries(ctx context.Context, startIndex, pageSize int,
	filters *database.Filters, sorters *database.Sorters,
) (
	int64, []*database.Entry, error,
) {

	// Check filters and sorters.
	//	TODO: implement support for multiple filters and sorters.
	var (
		fLen = filters.Len()
		sLen = sorters.Len()
	)
	if fLen > 1 {
		return 0, nil, errors.New(
			"only one filter is support as of now")
	}
	if sLen > 1 {
		return 0, nil, errors.New(
			"only one sorter is supported as of now")
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	args := []interface{}{EntryIndexName, "*"}

	// Filter.
	if fLen == 1 {
		filters.Range(func(key string, filter database.Filter) error {
			if key == "date" {
				key = "created"
			}
			if filter.Unit != "" {
				key = fmt.Sprintf("%s_%s", key, filter.Unit)
			}
			min := strconv.FormatInt(filter.Min, 10)
			max := strconv.FormatInt(filter.Max, 10)
			args = append(args, "FILTER", key, min, max)
			return nil
		})
	}

	// Sorter.
	if sLen == 1 {
		sorters.Range(func(_ int, sorter database.Sorter) error {
			var (
				direction, key = sorter.Extract()
				dirStr         string
			)
			switch direction {
			case database.SortAsc:
				dirStr = "ASC"
			case database.SortDesc:
				dirStr = "DESC"
			}
			args = append(args, "SORTBY", key, dirStr)
			return nil
		})
	}

	// Page size.
	args = append(args, "LIMIT", startIndex, pageSize)

	out, err := r.conn.Do("FT.SEARCH", args...)
	if err != nil {
		return 0, nil, err
	}

	data, ok := out.([]interface{})
	if !ok {
		return 0, nil, errors.New(
			"internal server error: results format is unexpected")
	} else if len(data) < 1 {
		return 0, nil, errors.New(
			"internal server error: results count is unexpected")
	}

	// Get count.
	count, err := redis.Int64(data[0], nil)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get results count")
	}

	// Get entries.
	entries := make([]*database.Entry, 0)
	for i := 2; i < len(data); i += 2 {
		var index = i/2 - 1
		dbEntry, ok := data[i].([]interface{})
		if !ok {
			fmt.Printf("[%d]: TYPE(%T)", index, data[i])
			return 0, nil, errors.Errorf(
				"internal server error: conversion of 'result -> []interface{}' at entries[%d]", index)
		}
		entryRaw, ok := dbEntry[len(dbEntry)-1].([]byte)
		if !ok {
			return 0, nil, errors.Errorf(
				"internal server error: conversion of 'interface{} -> []byte' at entries[%d]", index)
		}
		entry, err := database.EntryFromJson(entryRaw)
		if err != nil {
			return 0, nil, errors.Errorf(
				"failed to extract json entry at entries[%d]", index)
		}
		entries = append(entries, entry)
		fmt.Printf(">> INDEX('%d') VALUE('%v')\n", index, entries[index])
	}
	return count, entries, nil
}

func (r *RedisDB) SetReservationOfEntry(ctx context.Context, kittyID iko.KittyID, reservation string) (*database.Entry, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	key := EntryKey(kittyID)

	rawEntry, err := r.getRawEntry(key)
	if err != nil {
		return nil, err
	}
	entry, err := database.EntryFromJson(rawEntry)
	if err != nil {
		return nil, err
	}
	entry.Reservation = reservation

	args := []interface{}{
		EntryIndexName, key, "1.0",
		"REPLACE", "PARTIAL",
		"FIELDS",
		"reservation", reservation,
		"raw", entry.Json(),
	}

	if _, err := r.conn.Do("FT.ADD", args); err != nil {
		return nil, err
	}

	return entry, nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func (r *RedisDB) getRawEntry(key string) ([]byte, error) {
	out, err := redis.Values(r.conn.Do(
		"FT.GET", EntryIndexName, key))
	return redis.Bytes(out[len(out)-1], err)
}

func (r *RedisDB) addEntry(key string, raw []byte, entry *database.Entry) error {
	_, err := r.conn.Do("FT.ADD", r.getAddEntryArgs(key, raw, entry)...)
	return err
}

func (r *RedisDB) getAddEntryArgs(key string, raw []byte, entry *database.Entry) []interface{} {
	return []interface{}{
		EntryIndexName, key, "1.0",

		"FIELDS",

		"kitty_id", entry.ID,
		"name", entry.Name,
		"dna", entry.KittyDNA,

		"price_btc", entry.PriceBTC,
		"price_sky", entry.PriceSKY,
		"reservation", entry.Reservation,

		"created", time.Now().UnixNano(),
		"raw", raw,
	}
}

func entryKeys(entries []*database.Entry) []interface{} {
	out := make([]interface{}, len(entries))
	for i, entry := range entries {
		out[i] = EntryKey(entry.ID)
	}
	return out
}