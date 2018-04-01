package rpc

import (
	"net"
	"net/rpc"
	"sync"

	"github.com/kittycash/kitty-api/src/database"
	"github.com/skycoin/skycoin/src/cipher"
)

const (
	PrefixName  = "kc_api_rpc"
	NetworkName = "tcp"
)

type ServerConfig struct {
	Address   string // Address to serve on.
	TrustedPK string // Trusted public key for entry submissions.
}

func (sc *ServerConfig) ExtractTrustedPK() (cipher.PubKey, error) {
	return cipher.PubKeyFromHex(sc.TrustedPK)
}

type Server struct {
	rpc *rpc.Server
	lis net.Listener
	wg  sync.WaitGroup
}

func NewServer(c *ServerConfig, db database.Database) (*Server, error) {

	// Get trusted pk.
	pk, err := c.ExtractTrustedPK()
	if err != nil {
		return nil, err
	}

	// Prepare server.
	server := &Server{
		rpc: rpc.NewServer(),
	}
	if err := server.rpc.RegisterName(PrefixName, &Gateway{pk: pk, db: db}); err != nil {
		return nil, err
	}
	if server.lis, err = net.Listen(NetworkName, c.Address); err != nil {
		return nil, err
	}

	// Run service.
	if err := server.runService(); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *Server) runService() error {
	s.wg.Add(1)
	go func(l net.Listener) {
		defer s.wg.Done()
		s.rpc.Accept(l)
		// RPC Closes.
	}(s.lis)
	return nil
}

func (s *Server) Close() error {
	if err := s.lis.Close(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}
