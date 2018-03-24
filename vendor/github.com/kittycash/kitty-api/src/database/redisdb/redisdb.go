package redisdb

import (
	"context"
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/kittycash/kitty-api/src/database"
	"github.com/kittycash/wallet/src/iko"
	"github.com/pkg/errors"
	"time"
)

const (
	EntryPrefix   = "entry"
	EntryCountKey = "entry_count"
)

func EntryKey(id iko.KittyID) string {
	return fmt.Sprintf("%s:%d", EntryPrefix, id)
}

type Config struct {
	Address  string
	Database int
	Password string
}

type RedisDB struct {
	conn redis.Conn
	mux  sync.Mutex
}

func New(c *Config) (*RedisDB, error) {

	options := []redis.DialOption{
		redis.DialDatabase(c.Database),
		redis.DialConnectTimeout(time.Second * 10),
	}
	if c.Password != "" {
		options = append(options,
			redis.DialPassword(c.Password))
	}
	conn, err := redis.Dial("tcp", c.Address, options...)
	if err != nil {
		return nil, err
	}
	return &RedisDB{
		conn: conn,
	}, nil
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

	var key = EntryKey(entry.ID)
	fmt.Printf("[ADD] key(%v)\n", key)

	r.conn.Send("WATCH", key, EntryCountKey)
	r.conn.Send("MULTI")
	r.conn.Send("SET", key, entry.Json(), "NX")

	out, err := r.conn.Do("EXEC")
	if err != nil {
		return err
	}
	fmt.Println("Add[0]:", out)

	c, err := r.incCount()
	if err != nil {
		return err
	}
	fmt.Println("Add[1]:", c)

	return nil
}

func (r *RedisDB) Count(ctx context.Context) (int64, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.getCount()
}

func (r *RedisDB) GetEntryOfID(ctx context.Context, kittyID iko.KittyID) (*database.Entry, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	var key = EntryKey(kittyID)
	data, err := r.conn.Do("GET", key)
	if err != nil {
		return nil, err
	}

	switch data.(type) {
	case []byte:
		return database.EntryFromJson(data.([]byte))
	default:
		fmt.Sprintf("UNKNOWN TYPE! TYPE(%T) VALUE(%v)", data, data)
		return nil, errors.New("you screwed up")
	}
}

func (r *RedisDB) GetEntryOfDNA(ctx context.Context, kittyDNA string) (*database.Entry, error) {
	return nil, nil
}

func (r *RedisDB) GetEntries(ctx context.Context, startIndex, pageSize int,
	filters *database.Filters, sorters *database.Sorters) ([]*database.Entry, error) {

	return nil, nil
}

func (r *RedisDB) SetReservationOfEntry(ctx context.Context, kittyID iko.KittyID, isReserved bool) (*database.Entry, error) {
	return nil, nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func (r *RedisDB) getCount() (int64, error) {
	return redis.Int64(r.conn.Do("GET", EntryCountKey))
}

func (r *RedisDB) incCount() (int64, error) {
	c, err := r.conn.Do("INCR", EntryCountKey)
	if err != nil {
		return -1, err
	}
	v, _ := c.(int64)
	return v, nil
}
