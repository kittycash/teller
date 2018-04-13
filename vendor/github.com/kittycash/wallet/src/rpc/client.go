package rpc

import (
	"net/rpc"

	"gopkg.in/sirupsen/logrus.v1"
)

type ClientConfig struct {
	Address string
}

type Client struct {
	c   *ClientConfig
	l   *logrus.Logger
	rpc *rpc.Client
}

func NewClient(c *ClientConfig) (*Client, error) {
	var (
		e error
		s = &Client{
			c: c,
			l: logrus.New(),
		}
	)
	if s.rpc, e = rpc.Dial(NetworkName, c.Address); e != nil {
		return nil, e
	}
	return s, nil
}

func (c *Client) Close() {
	if c.rpc != nil {
		if e := c.rpc.Close(); e != nil {
			c.l.WithError(e).Error("error on rpc client close")
		} else {
			c.l.Info("rpc client closed")
		}
	}
}

func (c *Client) Balances(in *BalancesIn) (*BalancesOut, error) {
	var (
		out = new(BalancesOut)
		e   = c.rpc.Call(method("Balances"), in, out)
	)
	return out, e
}

func (c *Client) KittyOwner(in *KittyOwnerIn) (*KittyOwnerOut, error) {
	var (
		out = new(KittyOwnerOut)
		e   = c.rpc.Call(method("KittyOwner"), in, out)
	)
	return out, e
}

func (c *Client) Transaction(in *TransactionIn) (*TransactionOut, error) {
	var (
		out = new(TransactionOut)
		e   = c.rpc.Call(method("Transaction"), in, out)
	)
	return out, e
}

func (c *Client) InjectTx(in *InjectTxIn) (*InjectTxOut, error) {
	var (
		out = new(InjectTxOut)
		e   = c.rpc.Call(method("InjectTx"), in, out)
	)
	return out, e
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func method(v string) string {
	return PrefixName + "." + v
}

func empty() *struct{} {
	return &struct{}{}
}
