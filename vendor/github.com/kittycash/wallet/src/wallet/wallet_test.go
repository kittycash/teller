package wallet

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func initTempDir(t *testing.T) func() {
	dir, e := ioutil.TempDir("", "kittycash_test")
	require.Empty(t, e, "failed to create temp dir")

	e = SetRootDir(dir)
	require.Empty(t, e, "failed to set root dir")

	return func() {
		os.RemoveAll(dir)
	}
}

func saveWallet(t *testing.T, options *Options) {
	fWallet, e := NewFloatingWallet(options)
	require.Empty(t, e, "failed to create floating wallet")

	e = fWallet.Save()
	require.Empty(t, e, "failed to save wallet")
}

func loadWallet(t *testing.T, label, pw string) *Wallet {
	f, e := os.Open(LabelPath(label))
	require.Nilf(t, e, "failed to open wallet of label '%s'", label)
	defer f.Close()

	fw, e := LoadFloatingWallet(f, label, pw)
	require.Empty(t, e, "failed to load floating wallet")

	return fw
}

func TestFloatingWallet_Save(t *testing.T) {
	rmTemp := initTempDir(t)
	defer rmTemp()

	run := func(o *Options) {
		saveWallet(t, o)
		fw := loadWallet(t, o.Label, o.Password)
		m := fw.Meta
		require.Equal(t, m.Password, o.Password, "passwords do not match")
		require.Equal(t, m.Encrypted, o.Encrypted, "encrypted does not match")
		require.Equal(t, m.Label, o.Label, "label does not match")
		require.Equal(t, m.Seed, o.Seed, "seed does not match")
	}

	cases := []Options{
		{
			Label:     "wallet0",
			Seed:      "secure seed",
			Encrypted: true,
			Password:  "password",
		},
		{
			Label:     "wallet1",
			Seed:      "secure seed",
			Encrypted: false,
			Password:  "",
		},
	}

	for _, c := range cases {
		run(&c)
	}

}
