package iko

import "testing"

func TestKittyIDs_Sort(t *testing.T) {
	ids := KittyIDs{
		KittyID(65),
		KittyID(2),
		KittyID(20),
		KittyID(23),
		KittyID(12),
		KittyID(3),
		KittyID(94),
		KittyID(24),
	}
	ids.Sort()
	t.Log(ids)
}
