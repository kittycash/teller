package rpc

import "net/rpc"

type ClientConfig struct {
	Address string
}

type Client struct {
	c   *ClientConfig
	rpc *rpc.Client
}

func NewClient(c *ClientConfig) (*Client, error) {
	var (
		err    error
		client = &Client{
			c: c,
		}
	)
	if client.rpc, err = rpc.Dial(NetworkName, c.Address); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Close() error {
	if c.rpc != nil {
		if err := c.rpc.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) AddEntry(in *EntryIn) error {
	return c.rpc.Call(method("AddEntry"), in, new(struct{}))
}

func (c *Client) AddEntries(in *AddEntriesIn) error {
	return c.rpc.Call(method("AddEntries"), in, new(struct{}))
}

func (c *Client) Count() (*CountOut, error) {
	var (
		out = new(CountOut)
		err = c.rpc.Call(method("Count"), new(struct{}), out)
	)
	return out, err
}

func (c *Client) EntryOfID(in *EntryOfIDIn) (*EntryOfIDOut, error) {
	var (
		out = new(EntryOfIDOut)
		err = c.rpc.Call(method("EntryOfID"), in, out)
	)
	return out, err
}

func (c *Client) EntryOfDNA(in *EntryOfDNAIn) (*EntryOfDNAOut, error) {
	var (
		out = new(EntryOfDNAOut)
		err = c.rpc.Call(method("EntryOfDNA"), in, out)
	)
	return out, err
}

func (c *Client) Entries(in *EntriesIn) (*EntriesOut, error) {
	var (
		out = new(EntriesOut)
		err = c.rpc.Call(method("Entries"), in, out)
	)
	return out, err
}

func (c *Client) SetReservation(in *ReservationIn) (*ReservationOut, error) {
	var (
		out = new(ReservationOut)
		err = c.rpc.Call(method("SetReservation"), in, out)
	)
	return out, err
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func method(v string) string {
	return PrefixName + "." + v
}