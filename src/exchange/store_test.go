package exchange

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/util/testutil"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetBindAddress(depositAddr, coinType string) (*BoundAddress, error) {
	args := m.Called(depositAddr, coinType)

	ba := args.Get(0)
	if ba == nil {
		return nil, args.Error(1)
	}

	return ba.(*BoundAddress), args.Error(1)
}

func (m *MockStore) BindAddress(kittyID, depositAddr, coinType string) (*BoundAddress, error) {
	args := m.Called(kittyID, depositAddr, coinType)

	ba := args.Get(0)
	if ba == nil {
		return nil, args.Error(1)
	}

	return ba.(*BoundAddress), args.Error(1)
}

func (m *MockStore) GetOrCreateDepositInfo(dv scanner.Deposit) (DepositInfo, error) {
	args := m.Called(dv)
	return args.Get(0).(DepositInfo), args.Error(1)
}

func (m *MockStore) GetDepositInfoArray(filt DepositFilter) ([]DepositInfo, error) {
	args := m.Called(filt)

	dis := args.Get(0)
	if dis == nil {
		return nil, args.Error(1)
	}

	return dis.([]DepositInfo), args.Error(1)
}

func (m *MockStore) GetDepositInfoOfKittyID(kittyID string) ([]DepositInfo, error) {
	args := m.Called(kittyID)

	dis := args.Get(0)
	if dis == nil {
		return nil, args.Error(1)
	}

	return dis.([]DepositInfo), args.Error(1)
}

func (m *MockStore) UpdateDepositInfo(Tx string, f func(DepositInfo) DepositInfo) (DepositInfo, error) {
	args := m.Called(Tx, f)
	return args.Get(0).(DepositInfo), args.Error(1)
}

func (m *MockStore) UpdateDepositInfoCallback(btcTx string, f func(DepositInfo) DepositInfo, callback func(DepositInfo) error) (DepositInfo, error) {
	args := m.Called(btcTx, f, callback)
	return args.Get(0).(DepositInfo), args.Error(1)
}

func (m *MockStore) GetKittyBindAddress(kittyID string) (*BoundAddress, error) {
	args := m.Called(kittyID)

	bAddrs := args.Get(0)
	if bAddrs == nil {
		return nil, args.Error(1)
	}

	return bAddrs.(*BoundAddress), args.Error(1)
}

func (m *MockStore) GetDepositStats() (int64, int64, int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(int64), args.Error(2)
}

func newTestStore(t *testing.T) (*Store, func()) {
	db, shutdown := testutil.PrepareDB(t)

	log, _ := testutil.NewLogger(t)
	s, err := NewStore(log, db)
	require.NoError(t, err)

	return s, shutdown
}

func TestStoreNewStore(t *testing.T) {
	s, shutdown := newTestStore(t)
	defer shutdown()

	// check the buckets
	err := s.db.View(func(tx *bolt.Tx) error {
		require.NotNil(t, tx.Bucket(DepositInfoBkt))
		require.NotNil(t, tx.Bucket(MustGetBindAddressBkt(scanner.CoinTypeBTC)))
		require.NotNil(t, tx.Bucket(MustGetBindAddressBkt(scanner.CoinTypeSKY)))
		//@TODO (therealssj): Update buckets
		require.NotNil(t, tx.Bucket(BtcTxsBkt))
		return nil
	})
	require.NoError(t, err)
}

func mustBindAddress(t *testing.T, s Storer, skyAddr, addr string) {
	boundAddr, err := s.BindAddress(skyAddr, addr, scanner.CoinTypeBTC)
	require.NoError(t, err)
	require.NotNil(t, boundAddr)
	require.Equal(t, skyAddr, boundAddr.KittyID)
	require.Equal(t, addr, boundAddr.Address)
	require.Equal(t, scanner.CoinTypeBTC, boundAddr.CoinType)
	//@TODO (therealssj): update test
}

func TestStoreBindAddress(t *testing.T) {
	//@TODO (therealssj): update test
}

func TestStoreBindAddressTwiceFails(t *testing.T) {
	s, shutdown := newTestStore(t)
	defer shutdown()

	mustBindAddress(t, s, "a", "b")

	boundAddr, err := s.BindAddress("a", "b", scanner.CoinTypeBTC)
	require.Error(t, err)
	require.Equal(t, ErrAddressAlreadyBound, err)
	require.Nil(t, boundAddr)

	boundAddr, err = s.BindAddress("c", "b", scanner.CoinTypeBTC)
	require.Error(t, err)
	require.Equal(t, ErrAddressAlreadyBound, err)
	require.Nil(t, boundAddr)
}

//@TODO (therealssj): Add tests
