package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/scanner"
	"github.com/kittycash/teller/src/util/testutil"
)

type dummyBtcAddrMgr struct {
	Num uint64
}

func (db *dummyBtcAddrMgr) Remaining() uint64 {
	return db.Num
}

type dummySkyAddrMgr struct {
	Num uint64
}

func (db *dummySkyAddrMgr) Remaining() uint64 {
	return db.Num
}

type dummyDepositStatusGetter struct {
	dpis []exchange.DepositInfo
}

func (dps dummyDepositStatusGetter) GetDepositStatusDetail(flt exchange.DepositFilter) ([]exchange.DepositStatusDetail, error) {
	var ds []exchange.DepositStatusDetail
	for _, dpi := range dps.dpis {
		if flt(dpi) {
			ds = append(ds, exchange.DepositStatusDetail{
				Seq:            dpi.Seq,
				DepositAddress: dpi.DepositAddress,
				OwnerAddress:   dpi.OwnerAddress,
				Status:         dpi.Status.String(),
				UpdatedAt:      dpi.UpdatedAt,
				Txid:           dpi.Txid,
				CoinType:       dpi.CoinType,
			})
		}
	}
	return ds, nil
}

func (dps dummyDepositStatusGetter) GetDepositStats() (*exchange.DepositStats, error) {
	var totalBTCReceived int64
	var totalSKYReceived int64
	var totalBoxesSent int64
	for _, dpi := range dps.dpis {
		if dpi.CoinType == scanner.CoinTypeBTC {
			totalBTCReceived += dpi.DepositValue
		}

		if dpi.CoinType == scanner.CoinTypeSKY {
			totalSKYReceived += dpi.DepositValue
		}

		// TotalBoxesSent = no. of deposits with status == done
		if dpi.Status == exchange.StatusDone {
			totalBoxesSent++
		}

	}
	return &exchange.DepositStats{
		TotalBTCReceived: totalBTCReceived,
		TotalSKYReceived: totalSKYReceived,
		TotalBoxesSent:   totalBoxesSent,
	}, nil
}

type dummyScanAddrs struct {
	// addrs []string
}

func (ds dummyScanAddrs) GetScanAddresses() ([]string, error) {
	return []string{}, nil
}

func TestRunMonitor(t *testing.T) {
	dpis := []exchange.DepositInfo{
		{
			DepositAddress: "b1",
			OwnerAddress:   "s1",
			Status:         exchange.StatusWaitDeposit,
		},
		{
			DepositAddress: "b2",
			OwnerAddress:   "s2",
			Status:         exchange.StatusWaitSend,
		},
		{
			DepositAddress: "b3",
			OwnerAddress:   "s3",
			Status:         exchange.StatusWaitConfirm,
		},
		{
			DepositAddress: "b4",
			OwnerAddress:   "s4",
			Status:         exchange.StatusDone,
		},
		{
			DepositAddress: "b5",
			OwnerAddress:   "s6",
			Status:         exchange.StatusDone,
		},
	}

	dummyDps := dummyDepositStatusGetter{dpis: dpis}

	cfg := Config{
		"localhost:7908",
	}

	log, _ := testutil.NewLogger(t)
	m := New(log, cfg, &dummyBtcAddrMgr{10}, &dummySkyAddrMgr{10}, &dummyDps, &dummyScanAddrs{})

	time.AfterFunc(1*time.Second, func() {
		rsp, err := http.Get(fmt.Sprintf("http://localhost:7908/api/address"))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rsp.StatusCode)

		var addrUsage addressUsage
		err = json.NewDecoder(rsp.Body).Decode(&addrUsage)
		require.NoError(t, err)
		require.Equal(t, uint64(10), addrUsage.RestAddrNum)
		testutil.CheckError(t, rsp.Body.Close)

		var tt = []struct {
			name        string
			status      string
			expectCode  int //
			expectValue []exchange.DepositInfo
		}{
			{
				"get deposit that are in waiting_deposit status",
				"waiting_deposit",
				http.StatusOK,
				dpis[:1],
			},
			{
				"get deposit that are in waiting_send status",
				"waiting_send",
				http.StatusOK,
				dpis[1:2],
			},
			{
				"get deposit that are in waiting_confirm status",
				"waiting_confirm",
				http.StatusOK,
				dpis[2:3],
			},
			{
				"get deposit that are in waiting_done status",
				"done",
				http.StatusOK,
				dpis[3:5],
			},
			{
				"get unknown status",
				"invalid",
				http.StatusBadRequest,
				nil,
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				rsp, err := http.Get(fmt.Sprintf("http://localhost:7908/api/deposit_status?status=%s", tc.status))
				require.NoError(t, err)
				defer testutil.CheckError(t, rsp.Body.Close)
				require.Equal(t, tc.expectCode, rsp.StatusCode)

				if rsp.StatusCode == http.StatusOK {
					var st []exchange.DepositStatusDetail
					err := json.NewDecoder(rsp.Body).Decode(&st)
					require.NoError(t, err)

					dss := make([]exchange.DepositInfo, 0, len(st))
					for _, s := range st {
						dss = append(dss, exchange.DepositInfo{
							Seq:            s.Seq,
							UpdatedAt:      s.UpdatedAt,
							Status:         exchange.NewStatusFromStr(s.Status),
							DepositAddress: s.DepositAddress,
							OwnerAddress:   s.OwnerAddress,
							Txid:           s.Txid,
						})
					}
					require.Equal(t, tc.expectValue, dss)
				}
			})
		}

		m.Shutdown()
	})

	if err := m.Run(); err != nil {
		return
	}
}
