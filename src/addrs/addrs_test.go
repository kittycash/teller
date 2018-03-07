package addrs

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kittycash/teller/src/util/testutil"
)

func testNewBtcAddrManager(t *testing.T, db *bolt.DB, log *logrus.Logger) (*Addrs, []string) {
	addresses := []string{
		"14JwrdSxYXPxSi6crLKVwR4k2dbjfVZ3xj",
		"1JNonvXRyZvZ4ZJ9PE8voyo67UQN1TpoGy",
		"1JrzSx8a9FVHHCkUFLB2CHULpbz4dTz5Ap",
	}

	btca, err := NewAddrs(log, db, addresses, "test_bucket")
	require.NoError(t, err)

	addrMap := make(map[string]struct{}, len(btca.addresses))
	for _, a := range btca.addresses {
		addrMap[a] = struct{}{}
	}

	for _, addr := range addresses {
		_, ok := addrMap[addr]
		require.True(t, ok)
	}
	return btca, addresses
}

func testNewEthAddrManager(t *testing.T, db *bolt.DB, log *logrus.Logger) (*Addrs, []string) {
	addresses := []string{
		"0x12bc2e62a27f8940c373ef1edef7b615aeb045f3",
		"0x3e0081aa902a21ff8db61b29c05889a3d1b34f45",
		"0x50e0c87ef74079650ae6cd4ee895f8e1b02714cf",
	}

	etha, err := NewAddrs(log, db, addresses, "test_bucket_eth")
	require.NoError(t, err)

	addrMap := make(map[string]struct{}, len(etha.addresses))
	for _, a := range etha.addresses {
		addrMap[a] = struct{}{}
	}

	for _, addr := range addresses {
		_, ok := addrMap[addr]
		require.True(t, ok)
	}
	return etha, addresses
}

func testNewSkyAddrManager(t *testing.T, db *bolt.DB, log *logrus.Logger) (*Addrs, []string) {
	addresses := []string{
		"2Ag9SGMnVyaxzQbGL1EUfau2Fx1ztfNZsWt",
		"27XiEvCCEhB922y1r6JzYUN4AQA6XE6KnvL",
		"yW21jHJDWUtN8e5RwNtEgmQxVK14LCf4jS",
	}

	skya, err := NewAddrs(log, db, addresses, "test_bucket_sky")
	require.NoError(t, err)

	addrMap := make(map[string]struct{}, len(skya.addresses))
	for _, a := range skya.addresses {
		addrMap[a] = struct{}{}
	}

	for _, addr := range addresses {
		_, ok := addrMap[addr]
		require.True(t, ok)
	}

	return skya, addresses
}

func TestNewBtcAddrs(t *testing.T) {
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()

	log, _ := testutil.NewLogger(t)
	testNewBtcAddrManager(t, db, log)
}

func TestNewAddress(t *testing.T) {
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()

	addresses := []string{
		"14JwrdSxYXPxSi6crLKVwR4k2dbjfVZ3xj",
		"1JNonvXRyZvZ4ZJ9PE8voyo67UQN1TpoGy",
		"1JrzSx8a9FVHHCkUFLB2CHULpbz4dTz5Ap",
		"1JrzSx8a9FVHHCkUFLB2CHULpbz4dTz5Ap",
	}

	log, _ := testutil.NewLogger(t)
	btca, err := NewAddrs(log, db, addresses, "test_bucket")
	require.NoError(t, err)

	addr, err := btca.NewAddress()
	require.NoError(t, err)

	addrMap := make(map[string]struct{})
	for _, a := range btca.addresses {
		addrMap[a] = struct{}{}
	}

	// check if the addr still in the address pool
	_, ok := addrMap[addr]
	require.False(t, ok)

	// check if the addr is in used storage
	used, err := btca.used.IsUsed(addr)
	require.NoError(t, err)
	require.True(t, used)

	log, _ = testutil.NewLogger(t)
	btca1, err := NewAddrs(log, db, addresses, "test_bucket")
	require.NoError(t, err)

	for _, a := range btca1.addresses {
		require.NotEqual(t, a, addr)
	}

	used, err = btca1.used.IsUsed(addr)
	require.NoError(t, err)
	require.True(t, used)

	// run out all addresses
	for i := 0; i < 2; i++ {
		_, err = btca1.NewAddress()
		require.NoError(t, err)
	}

	_, err = btca1.NewAddress()
	require.Error(t, err)
	require.Equal(t, ErrDepositAddressEmpty, err)
}

func TestNewSkyAddrs(t *testing.T) {
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()
	log, _ := testutil.NewLogger(t)
	testNewSkyAddrManager(t, db, log)
}

func TestNewSkyAddress(t *testing.T) {
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()

	addresses := []string{
		"2Ag9SGMnVyaxzQbGL1EUfau2Fx1ztfNZsWt",
		"27XiEvCCEhB922y1r6JzYUN4AQA6XE6KnvL",
		"yW21jHJDWUtN8e5RwNtEgmQxVK14LCf4jS",
		"yW21jHJDWUtN8e5RwNtEgmQxVK14LCf4jS",
	}

	log, _ := testutil.NewLogger(t)
	skya, err := NewAddrs(log, db, addresses, "test_bucket_sky")
	require.NoError(t, err)

	addr, err := skya.NewAddress()
	require.NoError(t, err)

	addrMap := make(map[string]struct{})
	for _, a := range skya.addresses {
		addrMap[a] = struct{}{}
	}

	// check if the addr still in the address pool
	_, ok := addrMap[addr]
	require.False(t, ok)

	// check if the addr is in used storage
	used, err := skya.used.IsUsed(addr)
	require.NoError(t, err)
	require.True(t, used)

	log, _ = testutil.NewLogger(t)
	skya1, err := NewAddrs(log, db, addresses, "test_bucket_skya")
	require.NoError(t, err)

	for _, a := range skya1.addresses {
		require.NotEqual(t, a, addr)
	}

	used, err = skya1.used.IsUsed(addr)
	require.NoError(t, err)
	require.True(t, used)

	// run out all addresses
	for i := 0; i < 3; i++ {
		_, err = skya1.NewAddress()
		require.NoError(t, err)
	}

	_, err = skya1.NewAddress()
	require.Error(t, err)
	require.Equal(t, ErrDepositAddressEmpty, err)
}

func TestAddrManager(t *testing.T) {
	db, shutdown := testutil.PrepareDB(t)
	defer shutdown()
	log, _ := testutil.NewLogger(t)

	//create AddrGenertor
	btcGen, btcAddresses := testNewBtcAddrManager(t, db, log)
	ethGen, ethAddresses := testNewEthAddrManager(t, db, log)

	typeB := "TOKENB"
	typeE := "TOKENE"

	addrManager := NewAddrManager()
	//add generator to addrManager
	err := addrManager.PushGenerator(btcGen, typeB)
	require.NoError(t, err)
	err = addrManager.PushGenerator(ethGen, typeE)
	require.NoError(t, err)

	addrMap := make(map[string]struct{})
	for _, a := range btcAddresses {
		addrMap[a] = struct{}{}
	}
	// run out all addresses of typeB
	for i := 0; i < len(btcAddresses); i++ {
		addr, err := addrManager.NewAddress(typeB)
		require.NoError(t, err)
		//the addr still in the address pool
		_, ok := addrMap[addr]
		require.True(t, ok)
	}
	//the address pool of typeB is empty
	_, err = addrManager.NewAddress(typeB)
	require.Equal(t, ErrDepositAddressEmpty, err)

	//set typeE address into map
	addrMap = make(map[string]struct{})
	for _, a := range ethAddresses {
		addrMap[a] = struct{}{}
	}

	// run out all addresses of typeE
	for i := 0; i < len(ethAddresses); i++ {
		addr, err := addrManager.NewAddress(typeE)
		require.NoError(t, err)
		// check if the addr still in the address pool
		_, ok := addrMap[addr]
		require.True(t, ok)
	}
	_, err = addrManager.NewAddress(typeE)
	require.Equal(t, ErrDepositAddressEmpty, err)

	//check not exists cointype
	_, err = addrManager.NewAddress("OTHERTYPE")
	require.Equal(t, ErrCoinTypeNotExists, err)
}
