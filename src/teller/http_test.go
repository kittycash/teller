package teller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"errors"

	"github.com/kittycash/teller/src/exchange"
	"github.com/kittycash/teller/src/sender"
	"github.com/kittycash/teller/src/util/testutil"
)

type fakeExchanger struct {
	mock.Mock
}

func (e *fakeExchanger) BindAddress(kittyID, depositAddr, coinType string) (*exchange.BoundAddress, error) {
	args := e.Called(kittyID, depositAddr, coinType)

	ba := args.Get(0)
	if ba == nil {
		return nil, args.Error(1)
	}

	return ba.(*exchange.BoundAddress), args.Error(1)
}

func (e *fakeExchanger) GetDepositStatuses(kittyID string) ([]exchange.DepositStatus, error) {
	args := e.Called(kittyID)
	return args.Get(0).([]exchange.DepositStatus), args.Error(1)
}

func (e *fakeExchanger) GetDepositStatusDetail(flt exchange.DepositFilter) ([]exchange.DepositStatusDetail, error) {
	args := e.Called(flt)
	return args.Get(0).([]exchange.DepositStatusDetail), args.Error(1)
}

func (e *fakeExchanger) IsBound(kittyID string) bool {
	args := e.Called(kittyID)
	return args.Bool(0)
}

func (e *fakeExchanger) GetDepositStats() (*exchange.DepositStats, error) {
	args := e.Called()
	return args.Get(0).(*exchange.DepositStats), args.Error(1)
}

func (e *fakeExchanger) Status() error {
	args := e.Called()
	return args.Error(0)
}

func (e *fakeExchanger) Balance() int {
	//@TODO (therealssj): fix this

	args := e.Called()
	r := args.Get(0).(*int)

	return *r
}

func TestExchangeStatusHandler(t *testing.T) {
	tt := []struct {
		name           string
		method         string
		url            string
		status         int
		err            string
		exchangeStatus error
		errorMsg       string
		balance        int
		balanceError   error
	}{
		{
			"405",
			http.MethodPost,
			"/api/exchange-status",
			http.StatusMethodNotAllowed,
			"Invalid request method",
			nil,
			"",
			1,
			nil,
		},

		{
			"200",
			http.MethodGet,
			"/api/exchange-status",
			http.StatusOK,
			"",
			nil,
			"",
			1,
			nil,
		},

		{
			"200 status message error ignored, not APIError",
			http.MethodGet,
			"/api/exchange-status",
			http.StatusOK,
			"",
			errors.New("exchange.Status error"),
			"",
			1,
			nil,
		},

		{
			"200 status message error is APIError",
			http.MethodGet,
			"/api/exchange-status",
			http.StatusOK,
			"",
			sender.NewAPIError(errors.New("exchange.Status API error")),
			"exchange.Status API error",
			1,
			nil,
		},
		//@TODO (therealssj): add more tests when wallet is implemented properly

	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			e := &fakeExchanger{}

			e.On("Status").Return(tc.exchangeStatus)

			if tc.balanceError == nil {
				e.On("Balance").Return(&tc.balance, nil)
			} else {
				e.On("Balance").Return(nil, tc.balanceError)
			}

			req, err := http.NewRequest(tc.method, tc.url, nil)
			require.NoError(t, err)

			log, _ := testutil.NewLogger(t)

			rr := httptest.NewRecorder()
			httpServ := &HTTPServer{
				log:       log,
				exchanger: e,
			}
			handler := httpServ.setupMux()

			handler.ServeHTTP(rr, req)

			status := rr.Code
			require.Equal(t, tc.status, status, "wrong status code: got `%v` want `%v`", tc.name, status, tc.status)

			if status != http.StatusOK {
				require.Equal(t, tc.err, strings.TrimSpace(rr.Body.String()))
				return
			}

			var msg ExchangeStatusResponse
			err = json.Unmarshal(rr.Body.Bytes(), &msg)
			require.NoError(t, err)
			require.Equal(t, ExchangeStatusResponse{
				Error: tc.errorMsg,
				Balance: ExchangeStatusResponseBalance{
					Kitties: tc.balance,
				},
			}, msg)
		})
	}

}
