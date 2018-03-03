package iko

import (
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
	"testing"
)

func runStateDBTest(t *testing.T, stateDB StateDB) {
	t.Run("GetKittyState_NoKittiesAvailable", func(t *testing.T) {
		// since we don't have any kitties yet, GetKittyState should never give us a non-nil response at this point
		kittyState, kittyExists := stateDB.GetKittyState(KittyID(0))

		require.Nil(t, kittyState, "No kitties available yet")
		require.False(t, kittyExists, "No kitties available yet")
	})

	anAddress := cipher.AddressFromSecKey(
		cipher.SecKey([32]byte{
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
			3, 4, 5, 6,
		}))

	t.Run("GetAddressState_NoKittiesAvailable", func(t *testing.T) {
		// GetAddressState should always return a non-nil result
		addressState := stateDB.GetAddressState(anAddress)

		require.NotNil(t, addressState, "GetAddressState always returns a non-nil result")

		require.Len(t, addressState.Kitties, 0, "Address does not have any kitties yet")
		require.Len(t, addressState.Transactions, 0, "Address does not have any transactions yet either")
	})

	t.Run("AddKitty_Success", func(t *testing.T) {
		// let's add some kitties
		txHash := TxHash(cipher.SumSHA256([]byte{3, 4, 5, 6}))
		kID := KittyID(3)
		noSuchKID := KittyID(6)

		err := stateDB.AddKitty(txHash, kID, anAddress)

		require.Nil(t, err, "Adding our first kitty works")

		t.Run("AddKitty_Failure", func(t *testing.T) {
			// but trying to add that same kitty twice shouldn't work
			err = stateDB.AddKitty(txHash, kID, anAddress)

			require.NotNil(t, err, "Adding a kitty twice should fail")
		})

		t.Run("GetKittyState_Success", func(t *testing.T) {
			kittyState, kittyExists := stateDB.GetKittyState(kID)

			require.NotNil(t, kittyState, "Successfully fetched KittyState")
			require.True(t, kittyExists, "Successfully fetched KittyState")

			require.Equal(t, kittyState.Address, anAddress, "Address matches up")
			require.Equal(t, kittyState.Transactions, TxHashes{txHash}, "Transaction hashes match up")
		})

		t.Run("GetAddressState_Success", func(t *testing.T) {
			// in preparation, let's add another kitty
			secondTxHash := TxHash(cipher.SumSHA256([]byte{7, 8, 9, 10}))
			secondKID := KittyID(2)
			err := stateDB.AddKitty(secondTxHash, secondKID, anAddress)

			require.Nil(t, err, "Adding a second kitty should succeed")

			// GetAddressState should always return a non-nil result
			addressState := stateDB.GetAddressState(anAddress)

			require.NotNil(t, addressState, "GetAddressState always returns a non-nil result")

			require.Equal(t, addressState.Kitties, KittyIDs{secondKID, kID}, "Address should have two KittyIDs in ascending order")

			require.Contains(t, addressState.Transactions, txHash, "Address should have both transactions")
			require.Contains(t, addressState.Transactions, secondTxHash, "Address should have both transactions")
		})

		// now let's shuffle some kitties around
		secondTxHash := TxHash(cipher.SumSHA256([]byte{7, 8, 9, 10}))
		anotherAddress := cipher.AddressFromSecKey(
			cipher.SecKey([32]byte{
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,

				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
				7, 8, 9, 10,
			}))

		t.Run("MoveKitty_AlreadyOwned", func(t *testing.T) {
			err = stateDB.MoveKitty(secondTxHash, kID, anAddress, anAddress)

			require.NotNil(t, err, "You can't transfer a kitty to yourself")
		})

		t.Run("MoveKitty_KittyNapping", func(t *testing.T) {
			err = stateDB.MoveKitty(secondTxHash, kID, anotherAddress, anAddress)

			require.NotNil(t, err, "Kidnapping is not allowed")
		})

		t.Run("MoveKitty_NoSuchKitty", func(t *testing.T) {
			err = stateDB.MoveKitty(secondTxHash, noSuchKID, anAddress, anotherAddress)

			require.NotNil(t, err, "No such kitty")
		})

		t.Run("MoveKitty_Success", func(t *testing.T) {
			err = stateDB.MoveKitty(secondTxHash, kID, anAddress, anotherAddress)

			require.Nil(t, err, "Successfully transferred kitty")
		})
	})
}

func TestStateDB_MemoryState(t *testing.T) {
	stateDB := NewMemoryState()

	require.NotNil(t, stateDB, "We should be able to create an empty MemoryState")

	runStateDBTest(t, stateDB)
}
