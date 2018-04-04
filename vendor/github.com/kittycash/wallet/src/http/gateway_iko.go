package http

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"

	"github.com/kittycash/wallet/src/iko"
)

func ikoGateway(m *http.ServeMux, g *iko.BlockChain) error {
	Handle(m, "/api/iko/kitty/", "GET", getKitty(g))
	Handle(m, "/api/iko/address/", "GET", getAddress(g))
	Handle(m, "/api/iko/balance", "GET", getBalance(g))
	Handle(m, "/api/iko/tx/", "GET", getTx(g))
	Handle(m, "/api/iko/head_tx", "GET", getHeadTx(g))
	Handle(m, "/api/iko/txs", "GET", getPaginatedTxs(g))
	Handle(m, "/api/iko/inject_tx", "POST", injectTx(g))
	return nil
}

type KittyReply struct {
	KittyID      iko.KittyID `json:"kitty_id"`
	Address      string      `json:"address"`
	Transactions []string    `json:"transactions"`
}

func getKitty(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		kittyID, e := iko.KittyIDFromString(p.Base)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		kState, ok := g.GetKittyState(kittyID)
		if !ok {
			return sendJson(w, http.StatusNotFound,
				fmt.Sprintf("kitty of id '%d' not found", kittyID))
		}
		return SwitchTypeQuery(w, r, TqJson, TypeQueryActions{
			TqJson: func() error {
				return sendJson(w, http.StatusOK,
					KittyReply{
						KittyID:      kittyID,
						Address:      kState.Address.String(),
						Transactions: kState.Transactions.ToStringArray(),
					})
			},
			TqEnc: func() error {
				return sendBin(w, http.StatusOK,
					kState.Serialize())
			},
		})
	}
}

type AddressReply struct {
	Address      string       `json:"address"`
	Kitties      iko.KittyIDs `json:"kitties"`
	Transactions []string     `json:"transactions,omitempty"`
}

func getAddress(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		address, e := cipher.DecodeBase58Address(p.Base)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		aState := g.GetAddressState(address)
		return SwitchTypeQuery(w, r, TqJson, TypeQueryActions{
			TqJson: func() error {
				return sendJson(w, http.StatusOK,
					AddressReply{
						Address:      address.String(),
						Kitties:      aState.Kitties,
						Transactions: aState.Transactions.ToStringArray(),
					})
			},
			TqEnc: func() error {
				return sendBin(w, http.StatusOK,
					aState.Serialize())
			},
		})
	}
}

type BalanceReply struct {
	KittyCount int                     `json:"kitty_count"`
	Kitties    iko.KittyIDs            `json:"kitties"`
	PerAddress map[string]BalanceReply `json:"per_address,omitempty"`
}

func getBalance(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		var (
			qAddrs = r.URL.Query().Get("addrs")
		)
		addrs, e := toAddressArray(splitStr(qAddrs))
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("Error: %s", e.Error()))
		}
		var reply = BalanceReply{
			Kitties: make([]iko.KittyID, len(addrs)),
		}
		for _, addr := range addrs {
			aState := g.GetAddressState(addr)
			reply.KittyCount += len(aState.Kitties)
			reply.Kitties = append(reply.Kitties, aState.Kitties...)
		}
		reply.Kitties.Sort()
		return sendJson(w, http.StatusOK, reply)
	}
}

type TxMeta struct {
	Hash string `json:"hash"`
	Raw  string `json:"raw"`
	Seq  uint64 `json:"seq"`
	TS   int64  `json:"ts"`
}

type Tx struct {
	KittyID iko.KittyID `json:"kitty_id"`
	In      string      `json:"in"`
	Out     string      `json:"out"`
	Sig     string      `json:"sig"`
}

type TxReply struct {
	Meta TxMeta `json:"meta"`
	Tx   Tx     `json:"transaction"`
}

func NewTxReplyOfTransaction(txWrap iko.TxWrapper) TxReply {
	return TxReply{
		Meta: TxMeta{
			Hash: txWrap.Tx.Hash().Hex(),
			Raw:  hex.EncodeToString(txWrap.Tx.Serialize()),
			Seq:  txWrap.Meta.Seq,
			TS:   txWrap.Meta.TS,
		},
		Tx: Tx{
			KittyID: txWrap.Tx.KittyID,
			In:      txWrap.Tx.In.Hex(),
			Out:     txWrap.Tx.Out.String(),
			Sig:     txWrap.Tx.Sig.Hex(),
		},
	}
}

func getTx(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		var txWrap iko.TxWrapper
		ok, e := SwitchReqQuery(w, r, RqHash, ReqQueryActions{
			RqHash: func() (bool, error) {
				txHash, e := cipher.SHA256FromHex(p.Base)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						e.Error())
				}
				if txWrap, e = g.GetTxOfHash(iko.TxHash(txHash)); e != nil {
					return false, sendJson(w, http.StatusNotFound,
						e.Error())
				}
				return true, nil
			},
			RqSeq: func() (bool, error) {
				seq, e := strconv.ParseUint(p.Base, 10, 64)
				if e != nil {
					return false, sendJson(w, http.StatusBadRequest,
						e.Error())
				}
				if txWrap, e = g.GetTxOfSeq(seq); e != nil {
					return false, sendJson(w, http.StatusNotFound,
						e.Error())
				}
				return true, nil
			},
		})
		if !ok {
			return e
		}
		return SwitchTypeQuery(w, r, TqJson, TypeQueryActions{
			TqJson: func() error {
				return sendJson(w, http.StatusOK, NewTxReplyOfTransaction(txWrap))
			},
			TqEnc: func() error {
				return sendBin(w, http.StatusOK,
					txWrap.Tx.Serialize())
			},
		})
	}
}

func getHeadTx(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		txWrap, e := g.GetHeadTx()
		if e != nil {
			return sendJson(w, http.StatusNotFound,
				e.Error())
		}
		return SwitchTypeQuery(w, r, TqJson, TypeQueryActions{
			TqJson: func() error {
				return sendJson(w, http.StatusOK,
					NewTxReplyOfTransaction(txWrap))
			},
			TqEnc: func() error {
				return sendBin(w, http.StatusOK,
					txWrap.Tx.Serialize())
			},
		})
	}
}

type InjectTxRequest struct {
	Hex string `json:"hex"`
}

func injectTx(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		txRaw, e := ioutil.ReadAll(r.Body)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		var tx = new(iko.Transaction)
		switch contentType := r.Header.Get("Content-Type"); contentType {
		case "application/json":
			req := new(InjectTxRequest)
			if e := json.Unmarshal(txRaw, req); e != nil {
				return sendJson(w, http.StatusBadRequest,
					e.Error())
			}
			hexRaw, e := hex.DecodeString(req.Hex)
			if e != nil {
				return sendJson(w, http.StatusBadRequest,
					e.Error())
			}
			if e := encoder.DeserializeRaw(hexRaw, tx); e != nil {
				return sendJson(w, http.StatusBadRequest,
					e.Error())
			}
		case "application/octet-stream":
			if e := encoder.DeserializeRaw(txRaw, tx); e != nil {
				return sendJson(w, http.StatusBadRequest,
					e.Error())
			}
		default:
			return sendJson(w, http.StatusBadRequest,
				fmt.Sprintf("content type '%s' is not supported, expecting '%s'",
					contentType, []string{"application/json", "application/octet-stream"}))
		}

		meta, e := g.InjectTx(tx)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		return sendJson(w, http.StatusOK, meta)
	}
}

type PaginatedTxsReply struct {
	TotalPageCount uint64    `json:"total_page_count"`
	TxReplies      []TxReply `json:"transactions"`
}

func getPaginatedTxs(g *iko.BlockChain) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, p *Path) error {
		var (
			qCurrentPage = r.URL.Query().Get("current_page")
			qPerPage     = r.URL.Query().Get("per_page")
		)
		perPage, e := strconv.ParseUint(qPerPage, 10, 64)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		currentPage, e := strconv.ParseUint(qCurrentPage, 10, 64)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		paginated, e := g.GetTransactionPage(currentPage, perPage)
		if e != nil {
			return sendJson(w, http.StatusBadRequest,
				e.Error())
		}
		var txReplies []TxReply
		for _, transaction := range paginated.Transactions {
			txReplies = append(txReplies, NewTxReplyOfTransaction(transaction))
		}
		paginatedTxsReply := PaginatedTxsReply{
			TotalPageCount: paginated.TotalPageCount,
			TxReplies:      txReplies,
		}
		return sendJson(w, http.StatusOK,
			paginatedTxsReply)
	}
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func splitStr(in string) []string {
	out := strings.Split(in, ",")
	for i := len(out) - 1; i >= 0; i-- {
		if out[i] == "" {
			out = append(out[:i], out[i+1:]...)
		}
	}
	return out
}

func toAddressArray(in []string) ([]cipher.Address, error) {
	out := make([]cipher.Address, len(in))
	for i, vStr := range in {
		address, e := cipher.DecodeBase58Address(vStr)
		if e != nil {
			return nil, fmt.Errorf("invalid address '%s' at index '%d'",
				vStr, i)
		}
		out[i] = address
	}
	return out, nil
}
