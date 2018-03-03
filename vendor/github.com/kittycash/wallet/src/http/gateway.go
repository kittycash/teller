package http

import (
	"encoding/json"
	"fmt"
	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/wallet"
	"net/http"
	"path"
	"strings"
)

type Gateway struct {
	IKO    *iko.BlockChain
	Wallet *wallet.Manager
}

func (g *Gateway) host(mux *http.ServeMux) error {

	if g.IKO != nil {
		if e := ikoGateway(mux, g.IKO); e != nil {
			return e
		}
	}

	if g.Wallet != nil {
		if e := walletGateway(mux, g.Wallet); e != nil {
			return e
		}
	}

	return nil
}

/*
	<<< ACTION >>>
*/

type HandlerFunc func(w http.ResponseWriter, r *http.Request, p *Path) error

func Handle(mux *http.ServeMux, pattern, method string, handler HandlerFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {

		if r.Method != method {
			sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("invalid method type of '%s', expected '%s'",
					r.Method, method))

		} else if e := handler(w, r, NewPath(r)); e != nil {
			fmt.Println(e)
		}
	})
}

func MultiHandle(mux *http.ServeMux, patterns []string, method string, handler HandlerFunc) {
	for _, pattern := range patterns {
		Handle(mux, pattern, method, handler)
	}
}

func SwitchExtension(w http.ResponseWriter, p *Path, jsonAction, encAction func() error) error {
	switch p.Extension {
	case "", ".json":
		return jsonAction()
	case ".enc", ".bin":
		return encAction()
	default:
		return sendJson(w, http.StatusMethodNotAllowed,
			fmt.Sprintf("invalid URL extension '%s'", p.Extension))
	}
}

/*
	<<< RETURN SPECIFICATIONS >>>
*/

func sendJson(w http.ResponseWriter, status int, v interface{}) error {
	data, e := json.Marshal(v)
	if e != nil {
		return e
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, e = w.Write(data)
	return e
}

func sendBin(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(status)
	_, e := w.Write(data)
	return e
}

/*
	<<< URL Handler >>>
*/

type Path struct {
	EscapedPath string
	SplitPath   []string
	Extension   string
	Base        string
}

func NewPath(r *http.Request) *Path {
	var (
		escPath   = r.URL.EscapedPath()
		splitPath = strings.Split(escPath, "/")
		baseAll   = splitPath[len(splitPath)-1]
		ext       = path.Ext(baseAll)
		base      = strings.TrimSuffix(baseAll, ext)
	)
	return &Path{
		EscapedPath: escPath,
		SplitPath:   splitPath,
		Extension:   ext,
		Base:        base,
	}
}

func (p *Path) Segment(i int) string {
	if i < 0 || i >= len(p.SplitPath) {
		return ""
	}
	return p.SplitPath[i]
}
