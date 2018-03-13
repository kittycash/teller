package sender

import (
	"errors"

	"github.com/kittycash/wallet/src/iko"
)

var (
	// ErrSendBufferFull the Send service's request channel is full
	ErrSendBufferFull = errors.New("Send service's request queue is full")
	// ErrClosed the sender has closed
	ErrClosed = errors.New("Send service closed")
)

// Sender provids apis for sending skycoin
type Sender interface {
	CreateTransaction(recvAddr string, kittyID iko.KittyID) (*iko.Transaction, error)
	BroadcastTransaction(*iko.Transaction) *BroadcastTxResponse
	IsTxConfirmed(Txid string) *ConfirmResponse
	Balance(address string) (int, error)
}

// RetrySender provids helper function to send coins with Send service
// All requests will retry until succeeding.
type RetrySender struct {
	s *SendService
}

// NewRetrySender creates new sender
func NewRetrySender(s *SendService) *RetrySender {
	return &RetrySender{
		s: s,
	}
}

// CreateTransaction creates a transaction offline
func (s *RetrySender) CreateTransaction(recvAddr string, kittyID iko.KittyID) (*iko.Transaction, error) {
	return s.s.KittyClient.CreateTransaction(recvAddr, kittyID, s.s.secKey)
}

// BroadcastTransaction sends a transaction in a goroutine
func (s *RetrySender) BroadcastTransaction(tx *iko.Transaction) *BroadcastTxResponse {
	rspC := make(chan *BroadcastTxResponse, 1)

	go func() {
		s.s.broadcastTxChan <- BroadcastTxRequest{
			Tx:   tx,
			RspC: rspC,
		}
	}()

	return <-rspC
}

// IsTxConfirmed checks if tx is confirmed
func (s *RetrySender) IsTxConfirmed(txid string) *ConfirmResponse {
	rspC := make(chan *ConfirmResponse, 1)

	go func() {
		s.s.confirmChan <- ConfirmRequest{
			Txid: txid,
			RspC: rspC,
		}
	}()

	return <-rspC
}

// Balance returns the remaining balance of the sender
func (s *RetrySender) Balance(address string) (int, error) {
	return s.s.KittyClient.Balance(address)
}
