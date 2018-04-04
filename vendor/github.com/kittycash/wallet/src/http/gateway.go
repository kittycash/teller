package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kittycash/wallet/src/iko"
	"github.com/kittycash/wallet/src/wallet"
	"github.com/kittycash/wallet/src/kitties"
)

type Gateway struct {
	IKO    *iko.BlockChain
	Wallet *wallet.Manager
	Market *kitties.Manager
}

func (g *Gateway) host(mux *http.ServeMux) error {
	if g.IKO != nil {
		if e := ikoGateway(mux, g.IKO); e != nil {
			return e
		}
		if g.Market != nil {
			if e := marketKitties(mux, g.Market, g.IKO); e != nil {
				return e
			}
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

/*
	<<< CONTENT TYPE HEADER >>>
*/

type ContTypeVal string

const (
	ContTypeKey              = "Content-Type"
	CtApplicationJson        = ContTypeVal("application/json")
	CtApplicationOctetStream = ContTypeVal("application/octet-stream")
	CtApplicationForm        = ContTypeVal("application/x-www-form-urlencoded")
)

type ContTypeActions map[ContTypeVal]func() (bool, error)

func SwitchContType(w http.ResponseWriter, r *http.Request, m ContTypeActions) (bool, error) {
	v := ContTypeVal(r.Header.Get(ContTypeKey))
	action, ok := m[v]
	if !ok {
		return false, sendJson(w, http.StatusBadRequest,
			fmt.Sprintf("invalid '%s' query of '%s'", ReqQueryKey, v))
	}
	return action()
}

/*
	<<< REQUEST QUERY >>>
*/

type ReqQueryVal string

const (
	ReqQueryKey = "request"
	RqHash      = ReqQueryVal("hash")
	RqSeq       = ReqQueryVal("seq")
)

type ReqQueryActions map[ReqQueryVal]func() (bool, error)

func SwitchReqQuery(w http.ResponseWriter, r *http.Request, defVal ReqQueryVal, m ReqQueryActions) (bool, error) {
	v := ReqQueryVal(r.URL.Query().Get(ReqQueryKey))
	if v == "" {
		v = defVal
	}
	action, ok := m[v]
	if !ok {
		return false, sendJson(w, http.StatusBadRequest,
			fmt.Sprintf("invalid '%s' query of '%s'", ReqQueryKey, v))
	}
	return action()
}

/*
	<<< TYPE QUERY >>>
*/

type TypeQueryVal string

const (
	TypeQueryKey = "type"
	TqJson       = TypeQueryVal("json")
	TqEnc        = TypeQueryVal("enc")
)

type TypeQueryActions map[TypeQueryVal]func() error

func SwitchTypeQuery(w http.ResponseWriter, r *http.Request, defVal TypeQueryVal, m TypeQueryActions) error {
	v := TypeQueryVal(r.URL.Query().Get(TypeQueryKey))
	if v == "" {
		v = defVal
	}
	action, ok := m[v]
	if !ok {
		return sendJson(w, http.StatusBadRequest,
			fmt.Sprintf("invalid '%s' query of '%s'", TypeQueryKey, v))
	}
	return action()
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
	<<< TransformURL Handler >>>
*/

type Path struct {
	EscapedPath string
	SplitPath   []string
	Base        string
}

func NewPath(r *http.Request) *Path {
	var (
		escPath   = r.URL.EscapedPath()
		splitPath = strings.Split(escPath, "/")
		base      = splitPath[len(splitPath)-1]
	)
	return &Path{
		EscapedPath: escPath,
		SplitPath:   splitPath,
		Base:        base,
	}
}

func (p *Path) Segment(i int) string {
	if i < 0 || i >= len(p.SplitPath) {
		return ""
	}
	return p.SplitPath[i]
}
