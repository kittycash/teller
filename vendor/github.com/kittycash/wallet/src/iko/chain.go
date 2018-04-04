package iko

// TxChecker checks the transaction, returns an error when,
// there is a problem with the transaction, and it shouldn't
// be added to the blockchain.
type TxChecker func(tx *Transaction) error

// ChainDB represents where the transactions/blocks are stored.
// For iko, we combined blocks and transactions to become a single entity.
// Checks for whether txs are malformed shouldn't happen here.
type ChainDB interface {

	// Head should obtain the head transaction.
	// It should return an error when there are no transactions recorded.
	Head() (TxWrapper, error)

	// Len should obtain the length of the chain.
	Len() uint64

	// AddTx should add a transaction to the chain after the specified
	// 'check' returns nil.
	AddTx(txWrapper TxWrapper, check TxChecker) error

	// GetTxOfHash should obtain a transaction of a given hash.
	// It should return an error when the tx doesn't exist.
	GetTxOfHash(hash TxHash) (TxWrapper, error)

	// GetTxOfSeq should obtain a transaction of a given sequence.
	// It should return an error when the sequence given is invalid,
	//	or the tx doesn't exist.
	GetTxOfSeq(seq uint64) (TxWrapper, error)

	// TxChan obtains a channel where new transactions are sent through.
	// When a transaction is successfully saved to the `ChainDB` implementation,
	//	we expect to see it getting sent through here too.
	TxChan() <-chan *TxWrapper

	// GetTxsOfSeqRange returns a paginated portion of the Transactions.
	// It will return an error if the pageSize is zero
	// It will also return an error if startSeq is invalid
	GetTxsOfSeqRange(startSeq uint64, pageSize uint64) ([]TxWrapper, error)
}
