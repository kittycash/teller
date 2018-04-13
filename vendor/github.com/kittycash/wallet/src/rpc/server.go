package rpc

import (
	"net"
	"net/rpc"
	"sync"

	"gopkg.in/sirupsen/logrus.v1"
)

type ServerConfig struct {
	Address          string
	EnableRemoteQuit bool
}

type Server struct {
	c   *ServerConfig
	l   *logrus.Logger
	lis net.Listener
	rpc *rpc.Server
	api *Gateway
	wg  sync.WaitGroup
}

func NewServer(c *ServerConfig, g *Gateway) (*Server, error) {
	var (
		e error
		s = &Server{
			c:   c,
			l:   logrus.New(),
			rpc: rpc.NewServer(),
			api: g,
		}
	)
	if c.EnableRemoteQuit == false {
		s.api.QuitChan = nil
	}
	if e := s.rpc.RegisterName(PrefixName, s.api); e != nil {
		return nil, e
	}
	if s.lis, e = net.Listen("tcp", c.Address); e != nil {
		return nil, e
	}
	if e := s.runService(); e != nil {
		return nil, e
	}
	s.l.Infof("rpc listening on: '%s'", c.Address)
	return s, nil
}

func (s *Server) runService() error {
	s.wg.Add(1)
	go func(lis net.Listener) {
		defer s.wg.Done()
		s.rpc.Accept(lis)
		s.l.Print("rpc closed")
	}(s.lis)
	return nil
}

func (s *Server) Close() {
	if s.lis != nil {
		if e := s.lis.Close(); e != nil {
			s.l.WithError(e).Error("error on close")
		}
	}
	s.wg.Wait()
}
