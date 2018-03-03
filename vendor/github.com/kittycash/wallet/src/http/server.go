package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"
)

const (
	indexFileName = "index.html"
)

type ServerConfig struct {
	Address     string
	EnableGUI   bool
	GUIDir      string
	EnableTLS   bool
	TLSCertFile string
	TLSKeyFile  string
}

type Server struct {
	c    *ServerConfig
	srv  *http.Server
	mux  *http.ServeMux
	api  *Gateway
	quit chan struct{}
}

func NewServer(config *ServerConfig, api *Gateway) (*Server, error) {
	var server = &Server{
		c:    config,
		mux:  http.NewServeMux(),
		api:  api,
		quit: make(chan struct{}),
	}
	if e := server.prepareMux(); e != nil {
		return nil, e
	}
	go server.serve()
	return server, nil
}

func (s *Server) serve() {
	s.srv = &http.Server{
		Addr:    s.c.Address,
		Handler: s.mux,
	}
	if s.c.EnableTLS {
		for {
			if e := s.srv.ListenAndServeTLS(s.c.TLSCertFile, s.c.TLSKeyFile); e != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				break
			}
		}
	} else {
		for {
			if e := s.srv.ListenAndServe(); e != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				break
			}
		}
	}
	s.srv = nil
}

func (s *Server) prepareMux() error {
	if s.c.EnableGUI {
		if e := s.prepareGUI(); e != nil {
			return e
		}
	}
	return s.api.host(s.mux)
}

func (s *Server) prepareGUI() error {
	appLoc := s.c.GUIDir
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := path.Join(appLoc, indexFileName)
		http.ServeFile(w, r, page)
	})

	list, _ := ioutil.ReadDir(appLoc)
	for _, fInfo := range list {
		route := fmt.Sprintf("/%s", fInfo.Name())
		if fInfo.IsDir() {
			route += "/"
		}
		s.mux.Handle(route, http.FileServer(http.Dir(appLoc)))
	}
	return nil
}

// Close quits the http server.
func (s *Server) Close() {
	if s.quit != nil {
		close(s.quit)
		if s.srv != nil {
			s.srv.Close()
		}
	}
}
