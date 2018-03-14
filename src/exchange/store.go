package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/kittycash/teller/src/agent"
	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/util/dbutil"
)

var (
	// DepositInfoBkt maps a deposit transaction to a DepositInfo
	DepositInfoBkt = []byte("deposit_info")

	// TxsBkt maps a deposit transaction to a DepositInfo
	TxsBkt = []byte("deposit_txs")

	// KittyDepositSeqsIndexBkt maps a kitty id to its current deposit address
	KittyDepositSeqsIndexBkt = []byte("sky_deposit_seqs_index")

	//DepositTrackBkt keeps track of amount paid for a reservation box
	DepositTrackBkt = []byte("deposit_track")

	// ErrAddressAlreadyBound is returned if a payment address has already been bound to a kittyID
	ErrAddressAlreadyBound = errors.New("Address already bound to a kitty ID")
)

const bindAddressBktPrefix = "bind_address"

// GetBindAddressBkt returns the bind_address bucket name for a given coin type
func GetBindAddressBkt(coinType string) ([]byte, error) {
	var suffix string
	switch coinType {
	case scanner.CoinTypeBTC:
		suffix = "btc"
	case scanner.CoinTypeSKY:
		suffix = "sky"
	default:
		return nil, scanner.ErrUnsupportedCoinType
	}

	bktName := fmt.Sprintf("%s_%s", bindAddressBktPrefix, suffix)

	return []byte(bktName), nil
}

// MustGetBindAddressBkt panics if GetBindAddressBkt returns an error
func MustGetBindAddressBkt(coinType string) []byte {
	name, err := GetBindAddressBkt(coinType)
	if err != nil {
		panic(err)
	}
	return name
}

func init() {
	// Check that GetBindAddressBkt handles all possible coin types
	// TODO -- do similar init checks for other switches over coinType
	for _, ct := range scanner.GetCoinTypes() {
		name := MustGetBindAddressBkt(ct)
		if len(name) == 0 {
			panic(fmt.Sprintf("GetBindAddressBkt(%s) returned empty", ct))
		}
	}
}

// Storer interface for exchange storage
type Storer interface {
	GetBindAddress(depositAddr, coinType string) (*BoundAddress, error)
	BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error)
	GetOrCreateDepositInfo(scanner.Deposit) (DepositInfo, error)
	GetDepositInfoArray(DepositFilter) ([]DepositInfo, error)
	GetDepositInfoOfKittyID(string) ([]DepositInfo, error)
	UpdateDepositInfo(string, func(DepositInfo) DepositInfo) (DepositInfo, error)
	UpdateDepositInfoCallback(string, func(DepositInfo) DepositInfo, func(DepositInfo, *bolt.Tx) error) (DepositInfo, error)
	GetKittyBindAddress(string) (*BoundAddress, error)
	GetDepositStats() (int64, int64, int64, error)
	//TODO (therealssj): these need to be refactored
	getDepositTrack(depositAddr string) (DepositTrack, error)
	getDepositTrackTx(tx *bolt.Tx, depositAddr string) (DepositTrack, error)
	updateDepositTrack(depositAddr string, dt DepositTrack) error
	updateDepositTrackTx(tx *bolt.Tx, depositAddr string, dt DepositTrack) error
}

// Store storage for exchange
type Store struct {
	db  *bolt.DB
	log logrus.FieldLogger
}

// NewStore creates a Store instance
func NewStore(log logrus.FieldLogger, db *bolt.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("new exchange Store failed, db is nil")
	}

	if err := db.Update(func(tx *bolt.Tx) error {

		// create deposit status bucket if not exist
		if _, err := tx.CreateBucketIfNotExists(DepositInfoBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(DepositInfoBkt, err)
		}

		// create bind address bucket if not exist
		for _, ct := range scanner.GetCoinTypes() {
			bktName := MustGetBindAddressBkt(ct)
			if _, err := tx.CreateBucketIfNotExists(bktName); err != nil {
				return dbutil.NewCreateBucketFailedErr(bktName, err)
			}
		}

		if _, err := tx.CreateBucketIfNotExists(KittyDepositSeqsIndexBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(KittyDepositSeqsIndexBkt, err)
		}

		if _, err := tx.CreateBucketIfNotExists(TxsBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(TxsBkt, err)
		}

		if _, err := tx.CreateBucketIfNotExists(DepositTrackBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(DepositTrackBkt, err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &Store{
		db:  db,
		log: log.WithField("prefix", "exchange.Store"),
	}, nil
}

// GetBindAddress returns bound skycoin address of given bitcoin address.
// If no skycoin address is found, returns empty string and nil error.
func (s *Store) GetBindAddress(depositAddr, coinType string) (*BoundAddress, error) {
	var boundAddr *BoundAddress
	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		boundAddr, err = s.getBindAddressTx(tx, depositAddr, coinType)
		return err
	}); err != nil {
		return nil, err
	}

	return boundAddr, nil
}

// getBindAddressTx returns bound info of given deposit address.
func (s *Store) getBindAddressTx(tx *bolt.Tx, depositAddr, coinType string) (*BoundAddress, error) {
	bindBktFullName, err := GetBindAddressBkt(coinType)
	if err != nil {
		return nil, err
	}

	var boundAddr BoundAddress
	err = dbutil.GetBucketObject(tx, bindBktFullName, depositAddr, &boundAddr)
	switch err.(type) {
	case nil:
		return &boundAddr, nil
	case dbutil.ObjectNotExistErr:
		return nil, nil
	default:
		return nil, err
	}
}

// BindAddress binds a deposit address to bound info
func (s *Store) BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	log := s.log.WithField("kittyID", kittyID)
	log = log.WithField("depositAddr", depositAddr)
	log = log.WithField("coinType", coinType)

	bindBktFullName, err := GetBindAddressBkt(coinType)
	if err != nil {
		return nil, err
	}

	boundAddr := BoundAddress{
		KittyID:  kittyID,
		Address:  depositAddr,
		CoinType: coinType,
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		existingKittyID, err := s.getBindAddressTx(tx, depositAddr, coinType)
		if err != nil {
			return err
		}

		if existingKittyID != nil {
			err := ErrAddressAlreadyBound
			log.WithError(err).Error("Attempted to bind a payment address twice")
			return err
		}

		return dbutil.PutBucketValue(tx, bindBktFullName, depositAddr, boundAddr)
	}); err != nil {
		return nil, err
	}

	return &boundAddr, nil
}

// GetOrCreateDepositInfo creates a DepositInfo unless one exists with the DepositInfo.DepositID key,
// in which case it returns the existing DepositInfo.
func (s *Store) GetOrCreateDepositInfo(dv scanner.Deposit) (DepositInfo, error) {
	log := s.log.WithField("deposit", dv)

	var finalDepositInfo DepositInfo
	if err := s.db.Update(func(tx *bolt.Tx) error {
		di, err := s.getDepositInfoTx(tx, dv.ID())
		switch err.(type) {
		case nil:
			finalDepositInfo = di
			return nil

		case dbutil.ObjectNotExistErr:
			log.Info("DepositInfo not found in DB, inserting")
			boundAddr, err := s.getBindAddressTx(tx, dv.Address, dv.CoinType)
			if err != nil {
				err = fmt.Errorf("GetBindAddress failed: %v", err)
				log.WithError(err).Error(err)
				return err
			}

			if boundAddr == nil {
				err = ErrNoBoundAddress
				log.WithError(err).Error(err)
				return err
			}

			err = s.createDepositTrackTx(tx, boundAddr)
			if err != nil {
				err = fmt.Errorf("CreateDepositTrack failed: %v", err)
				log.WithError(err).Error(err)
				return err
			}

			log = log.WithField("boundAddr", boundAddr)

			// Integrity check of the boundAddr data against the deposit value data
			if boundAddr.CoinType != dv.CoinType {
				err := fmt.Errorf("boundAddr.CoinType != dv.CoinType")
				log.WithError(err).Error()
				return err
			}

			//TODO (therealssj): add owner address?
			di := DepositInfo{
				CoinType:       dv.CoinType,
				DepositAddress: dv.Address,
				KittyID:        boundAddr.KittyID,
				DepositID:      dv.ID(),
				Status:         StatusWaitDecide,
				DepositValue:   dv.Value,
				Deposit:        dv,
			}

			log = log.WithField("depositInfo", di)

			updatedDi, err := s.addDepositInfoTx(tx, di)
			if err != nil {
				err = fmt.Errorf("addDepositInfoTx failed: %v", err)
				log.WithError(err).Error(err)
				return err
			}

			finalDepositInfo = updatedDi

			return nil

		default:
			err = fmt.Errorf("getDepositInfo failed: %v", err)
			log.WithError(err).Error(err)
			return err
		}
	}); err != nil {
		return DepositInfo{}, err
	}

	return finalDepositInfo, nil

}

// addDepositInfo adds deposit info into storage, return seq or error
func (s *Store) addDepositInfo(di DepositInfo) (DepositInfo, error) {
	var updatedDi DepositInfo
	if err := s.db.Update(func(tx *bolt.Tx) error {
		var err error
		updatedDi, err = s.addDepositInfoTx(tx, di)
		return err
	}); err != nil {
		return di, err
	}

	return updatedDi, nil
}

// addDepositInfoTx adds deposit info into storage, return seq or error
func (s *Store) addDepositInfoTx(tx *bolt.Tx, di DepositInfo) (DepositInfo, error) {
	log := s.log.WithField("depositInfo", di)

	// check if the dpi with DepositID already exist
	if hasKey, err := dbutil.BucketHasKey(tx, DepositInfoBkt, di.DepositID); err != nil {
		return di, err
	} else if hasKey {
		return di, fmt.Errorf("deposit info of btctx \"%s\" already exists", di.DepositID)
	}

	seq, err := dbutil.NextSequence(tx, DepositInfoBkt)
	if err != nil {
		return di, err
	}

	updatedDi := di
	updatedDi.Seq = seq
	updatedDi.UpdatedAt = time.Now().UTC().Unix()

	if err := updatedDi.ValidateForStatus(); err != nil {
		log.WithError(err).Error("FIXME: Constructed invalid DepositInfo")
		return di, err
	}

	if err := dbutil.PutBucketValue(tx, DepositInfoBkt, updatedDi.DepositID, updatedDi); err != nil {
		return di, err
	}

	// update txs bucket
	var txs []string
	if err := dbutil.GetBucketObject(tx, TxsBkt, updatedDi.DepositAddress, &txs); err != nil {
		switch err.(type) {
		case dbutil.ObjectNotExistErr:
		default:
			return di, err
		}
	}

	txs = append(txs, updatedDi.DepositID)
	if err := dbutil.PutBucketValue(tx, TxsBkt, updatedDi.DepositAddress, txs); err != nil {
		return di, err
	}

	return updatedDi, nil
}

// getDepositInfo returns depsoit info of given address
func (s *Store) getDepositInfo(Txid string) (DepositInfo, error) {
	var di DepositInfo

	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		di, err = s.getDepositInfoTx(tx, Txid)
		return err
	})

	return di, err
}

// getDepositInfoTx returns depsoit info of given address
func (s *Store) getDepositInfoTx(tx *bolt.Tx, Txid string) (DepositInfo, error) {
	var dpi DepositInfo

	if err := dbutil.GetBucketObject(tx, DepositInfoBkt, Txid, &dpi); err != nil {
		return DepositInfo{}, err
	}

	return dpi, nil
}

// createDepositTrackTx creates a deposit track
func (s *Store) createDepositTrackTx(tx *bolt.Tx, boundInfo *BoundAddress) error {
	// check if the dpt with DepositAddr already exist
	if hasKey, err := dbutil.BucketHasKey(tx, DepositTrackBkt, boundInfo.Address); err != nil {
		return err
	} else if hasKey {
		return nil
	}

	price, err := s.getKittyPriceTx(tx, boundInfo.KittyID, boundInfo.CoinType)
	if err != nil {
		return err
	}

	dt := DepositTrack{
		KittyID:         boundInfo.KittyID,
		AmountDeposited: 0,
		AmountRequired:  price,
	}

	return dbutil.PutBucketValue(tx, DepositTrackBkt, boundInfo.Address, dt)
}

func (s *Store) updateDepositTrack(depositAddr string, dt DepositTrack) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.updateDepositTrackTx(tx, depositAddr, dt)
	})
}

// updateDepositTrackTx updates a deposit track
func (s *Store) updateDepositTrackTx(tx *bolt.Tx, depositAddr string, dt DepositTrack) error {
	return dbutil.PutBucketValue(tx, DepositTrackBkt, depositAddr, dt)
}

// getDepositTrack returns depsoit track of given address
func (s *Store) getDepositTrack(depositAddr string) (DepositTrack, error) {
	var dt DepositTrack

	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		dt, err = s.getDepositTrackTx(tx, depositAddr)
		return err
	})

	return dt, err
}

// getDepositTrackTx returns depsoit track of given address
func (s *Store) getDepositTrackTx(tx *bolt.Tx, depositAddr string) (DepositTrack, error) {
	var dpt DepositTrack

	if err := dbutil.GetBucketObject(tx, DepositTrackBkt, depositAddr, &dpt); err != nil {
		return DepositTrack{}, err
	}

	return dpt, nil
}

// GetDepositInfoArray returns filtered deposit info
func (s *Store) GetDepositInfoArray(flt DepositFilter) ([]DepositInfo, error) {
	var dpis []DepositInfo

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.ForEach(tx, DepositInfoBkt, func(k, v []byte) error {
			var dpi DepositInfo
			if err := json.Unmarshal(v, &dpi); err != nil {
				return err
			}

			if flt(dpi) {
				dpis = append(dpis, dpi)
			}

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return dpis, nil
}

// GetDepositInfoOfKittyID returns all deposit info that are bound
// to the given kittyID
func (s *Store) GetDepositInfoOfKittyID(kittyID string) ([]DepositInfo, error) {
	var dpis []DepositInfo

	if err := s.db.View(func(tx *bolt.Tx) error {
		boundAddr, err := s.GetKittyBindAddress(kittyID)
		if err != nil {
			return err
		}

		var txns []string
		if err := dbutil.GetBucketObject(tx, TxsBkt, boundAddr.Address, &txns); err != nil {
			switch err.(type) {
			case dbutil.ObjectNotExistErr:
			default:
				return err
			}
		}

		if len(txns) == 0 {
			dpis = append(dpis, DepositInfo{
				Status:         StatusWaitDeposit,
				DepositAddress: boundAddr.Address,
				KittyID:        kittyID,
				UpdatedAt:      time.Now().UTC().Unix(),
				CoinType:       boundAddr.CoinType,
			})
		}

		for _, txn := range txns {
			var dpi DepositInfo
			if err := dbutil.GetBucketObject(tx, DepositInfoBkt, txn, &dpi); err != nil {
				return err
			}

			dpis = append(dpis, dpi)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// sort the dpis by update time
	sort.Slice(dpis, func(i, j int) bool {
		return dpis[i].UpdatedAt < dpis[j].UpdatedAt
	})

	// renumber the seqs in the dpis
	for i := range dpis {
		dpis[i].Seq = uint64(i)
	}

	return dpis, nil
}

// UpdateDepositInfo updates deposit info. The update func takes a DepositInfo
// and returns a modified copy of it.
func (s *Store) UpdateDepositInfo(Tx string, update func(DepositInfo) DepositInfo) (DepositInfo, error) {
	return s.UpdateDepositInfoCallback(Tx, update, func(di DepositInfo, tx *bolt.Tx) error { return nil })
}

// UpdateDepositInfoCallback updates deposit info. The update func takes a DepositInfo
// and returns a modified copy of it.  After updating the DepositInfo, it calls callback,
// inside of the transaction.  If the callback returns an error, the DepositInfo update
// is rolled back.
func (s *Store) UpdateDepositInfoCallback(Txid string, update func(DepositInfo) DepositInfo, callback func(DepositInfo, *bolt.Tx) error) (DepositInfo, error) {
	log := s.log.WithField("Txid:", Txid)

	var dpi DepositInfo
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if err := dbutil.GetBucketObject(tx, DepositInfoBkt, Txid, &dpi); err != nil {
			return err
		}

		log = log.WithField("depositInfo", dpi)

		if dpi.DepositID != Txid {
			log.Error("DepositInfo.DepositID does not match Txid")
			err := fmt.Errorf("DepositInfo %+v saved under different key %s", dpi, Txid)
			return err
		}

		dpi = update(dpi)

		dpi.UpdatedAt = time.Now().UTC().Unix()

		if err := dbutil.PutBucketValue(tx, DepositInfoBkt, Txid, dpi); err != nil {
			return err
		}

		return callback(dpi, tx)

	}); err != nil {
		return DepositInfo{}, err
	}

	return dpi, nil
}

// GetKittyBindAddress returns the current bound address for a given kitty ID
func (s *Store) GetKittyBindAddress(kittyID string) (*BoundAddress, error) {
	// @TODO: improve this
	var boundAddr BoundAddress

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.GetBucketObject(tx, KittyDepositSeqsIndexBkt, kittyID, &boundAddr)
	}); err != nil {
		return nil, err
	}

	return &boundAddr, nil
}

func (s *Store) getKittyPriceTx(tx *bolt.Tx, kittyID string, coinType string) (int64, error) {
	var r agent.Reservation
	err := dbutil.GetBucketObject(tx, agent.ReservationsKittyBkt, kittyID, &r)
	if err != nil {
		return 0, err
	}

	if coinType == "SKY" {
		return r.Box.Detail.PriceSKY, nil
	} else if coinType == "BTC" {
		return r.Box.Detail.PriceBTC, nil
	}

	return 0, agent.ErrInvalidCoinType
}

// GetDepositStats returns BTC and SKY received and boxes sent
func (s *Store) GetDepositStats() (int64, int64, int64, error) {
	var totalBTCReceived int64
	var totalSKYReceived int64
	var totalBoxesSent int64

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.ForEach(tx, DepositInfoBkt, func(k, v []byte) error {
			var dpi DepositInfo
			if err := json.Unmarshal(v, &dpi); err != nil {
				return err
			}

			if dpi.CoinType == scanner.CoinTypeBTC {
				totalBTCReceived += dpi.DepositValue
			}

			if dpi.CoinType == scanner.CoinTypeSKY {
				totalSKYReceived += dpi.DepositValue
			}

			// TotalBoxesSent = no. of deposits with status == done
			if dpi.Status == StatusDone {
				totalBoxesSent++
			}

			return nil
		})
	}); err != nil {
		return -1, -1, -1, err
	}

	return totalBTCReceived, totalSKYReceived, totalBoxesSent, nil
}
