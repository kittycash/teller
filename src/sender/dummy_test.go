package sender


import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kittycash/wallet/src/iko"
	"github.com/skycoin/teller/src/util/testutil"
)

func TestDummySender(t *testing.T) {
	log, _ := testutil.NewLogger(t)

	s := NewDummySender(log)

	addr := "2VZu3rZozQ6nN37YSdj3EZJV7wSFVuLSm2X"
	var kittyID iko.KittyID = 9

	txn, err := s.CreateTransaction(addr, kittyID)
	require.NoError(t, err)
	require.NotNil(t, txn)

	// Another txn with the same dest addr and coins should have a different txid
	txn2, err := s.CreateTransaction(addr, kittyID)
	require.NoError(t, err)
	require.NotEqual(t, txn.Hash().Hex(), txn2.Hash().Hex())

	bRsp := s.BroadcastTransaction(txn)
	require.NotNil(t, bRsp)
	require.NoError(t, bRsp.Err)
	require.Equal(t, txn.Hash().Hex(), bRsp.Txid)

	// Broadcasting twice causes an error
	bRsp = s.BroadcastTransaction(txn)
	require.NotNil(t, bRsp)
	require.Error(t, bRsp.Err)
	require.Empty(t, bRsp.Txid)

	txHash := txn.Hash()
	cRsp := s.IsTxConfirmed(&txHash)
	require.NotNil(t, cRsp)
	require.NoError(t, cRsp.Err)
	require.False(t, cRsp.Confirmed)

	s.broadcastTxns[txn.Hash().Hex()].Confirmed = true

	txHash = txn.Hash()
	cRsp = s.IsTxConfirmed(&txHash)
	require.NotNil(t, cRsp)
	require.NoError(t, cRsp.Err)
	require.True(t, cRsp.Confirmed)
}
