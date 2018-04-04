package wallet

import (
	"errors"
	"io"
	"os"
	"sort"
	"sync"
)

var (
	ErrWalletNotFound     = errors.New("wallet of label is not found")
	ErrWalletLocked       = errors.New("wallet is locked")
	ErrLabelAlreadyExists = errors.New("label already exists")
)

// Manager manages the wallet files.
type Manager struct {
	mux     sync.Mutex
	labels  []string
	wallets map[string]*Wallet
}

// NewManager creates a new wallet manager.
func NewManager() (*Manager, error) {
	m := new(Manager)
	if e := m.Refresh(); e != nil {
		return nil, e
	}
	return m, nil
}

// Refresh reloads the list of wallets.
// All wallets will be locked.
func (m *Manager) Refresh() error {
	defer m.lock()()

	m.labels = make([]string, 0)
	m.wallets = make(map[string]*Wallet)
	e := RangeLabels(func(f io.Reader, label, fPath string, prefix Prefix) {
		if prefix.Version() != Version {
			log.Warningf(
				"wallet file `%s` is of version %v, while only version %v is supported",
				label, prefix.Version(), Version)
			return
		}
		var wallet *Wallet
		if prefix.Encrypted() {
			var e error
			if wallet, e = LoadFloatingWallet(f, label, ""); e != nil {
				return
			}
		}
		m.append(label, wallet)
	})
	if e != nil {
		return e
	}
	return m.sort()
}

// Stat represents a wallet when listed by 'ListWallets'.
type Stat struct {
	Label     string `json:"label"`
	Encrypted bool   `json:"encrypted"`
	Locked    *bool  `json:"locked,omitempty"`
}

// Lists the wallets available.
func (m *Manager) ListWallets() []Stat {
	defer m.lock()()

	var out = make([]Stat, len(m.labels))
	for i, label := range m.labels {
		fw := m.wallets[label]
		var (
			encrypted bool
			locked    *bool
		)
		if fw == nil {
			encrypted = true
			locked = new(bool)
			*locked = true
		} else {
			encrypted = fw.Meta.Encrypted
			if encrypted {
				locked = new(bool)
				*locked = false
			}
		}
		out[i] = Stat{
			Label:     label,
			Encrypted: encrypted,
			Locked:    locked,
		}
	}
	return out
}

// NewWallet creates a new wallet (and it's associated file)
// with specified options, and the number of addresses to generate under it.
func (m *Manager) NewWallet(opts *Options, addresses int) error {
	defer m.lock()()

	if addresses < 0 {
		return errors.New("can not have negative number of entries")
	}

	if _, ok := m.wallets[opts.Label]; ok {
		return ErrLabelAlreadyExists
	}

	fw, e := NewFloatingWallet(opts)
	if e != nil {
		return e
	}
	if e := fw.EnsureEntries(addresses); e != nil {
		return e
	}
	if e := fw.Save(); e != nil {
		return e
	}
	m.append(opts.Label, fw)
	return m.sort()
}

// DeleteWallet deletes a wallet of a given label.
func (m *Manager) DeleteWallet(label string) error {
	defer m.lock()()

	if m.remove(label) {
		return os.Remove(LabelPath(label))
	}
	return ErrWalletNotFound
}

// DisplayWallet displays the wallet of specified label.
// Password needs to be given if a wallet is still locked.
// Addresses ensures that wallet has at least the number of address entries.
func (m *Manager) DisplayWallet(label, password string, addresses int) (*FloatingWallet, error) {
	defer m.lock()()

	switch w, e := m.getWallet(label); e {
	case nil:
		if e := w.EnsureEntries(addresses); e != nil {
			return nil, e
		}
		return w.ToFloating(), nil

	case ErrWalletNotFound:
		return nil, ErrWalletNotFound

	case ErrWalletLocked:
		f, e := os.Open(LabelPath(label))
		if e != nil {
			return nil, e
		}
		defer f.Close()
		if w, e = LoadFloatingWallet(f, label, password); e != nil {
			return nil, e
		}
		m.wallets[label] = w
		if e := w.EnsureEntries(addresses); e != nil {
			return nil, e
		}
		return w.ToFloating(), nil

	default:
		return nil, errors.New("unknown error")
	}
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func (m *Manager) lock() func() {
	m.mux.Lock()
	return m.mux.Unlock
}

func (m *Manager) append(label string, fw *Wallet) {
	m.labels = append(m.labels, label)
	m.wallets[label] = fw
}

func (m *Manager) remove(label string) bool {
	for i, l := range m.labels {
		if l == label {
			m.labels = append(m.labels[:i], m.labels[i+1:]...)
			delete(m.wallets, label)
			return true
		}
	}
	return false
}

func (m *Manager) sort() error {
	sort.Strings(m.labels)
	return nil
}

func (m *Manager) getWallet(label string) (*Wallet, error) {
	w, ok := m.wallets[label]
	if !ok {
		return nil, ErrWalletNotFound
	}
	if w == nil {
		return nil, ErrWalletLocked
	}
	return w, nil
}
