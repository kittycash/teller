package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"time"

	"github.com/boltdb/bolt"
	"github.com/skycoin/teller/src/daemon"
	"github.com/skycoin/teller/src/logger"
	"github.com/skycoin/teller/src/service/cli"
	gconfig "github.com/skycoin/teller/src/service/config"
	"github.com/skycoin/teller/src/service/monitor"
	"github.com/skycoin/teller/src/service/scanner"
	"github.com/skycoin/teller/src/service/sender"
)

var (
	coinValueBktName      = []byte("coinValue")
	exchangeLogBktName    = []byte("exchangeLog")
	unconfirmedTxsBktName = []byte("unconfirmed_txs")
)

// skySender provids apis for sending skycoin
type skySender interface {
	SendAsync(destAddr string, coins int64, opt *sender.SendOption) (<-chan interface{}, error)
	Send(destAddr string, coins int64, opt *sender.SendOption) (string, error)
}

// skyScanner provids apis for interact with scan service
type btcScanner interface {
	AddDepositAddress(addr string) error
	GetDepositValue() <-chan scanner.DepositValue
}

type exchange struct {
	*exchgConfig
	scanner          btcScanner // scanner provides apis to interact with scan service
	sender           skySender  // sender provides apis to send skycoin
	coinValue        *coinValueBucket
	exchangeLogs     *exchangeLogBucket
	unconfirmedTxids *unconfirmedTxids
	cli              *cli.Cli
	quit             chan struct{}
}

type coinValue struct {
	Address    string // deposit coin address, like BTC
	CoinName   string // deposit coin name
	Balance    int64  // the balance of the deposit coin
	ICOAddress string // the ico coin address
}

type rateTable struct {
	rates []struct {
		Date time.Time
		Rate float64
	}
}

type exchgConfig struct {
	db          *bolt.DB
	log         logger.Logger
	rateTable   rateTable
	checkPeriod time.Duration // scan period
	nodeRPCAddr string        // the rpc address of node
	nodeWltFile string        // the fullpath of specific wallet file
	depositCoin string        // deposit coin name
	icoCoin     string        // ico coin name
}

func newExchange(cfg *exchgConfig) *exchange {
	ex := &exchange{
		coinValue:        newCoinValueBucket(coinValueBktName, cfg.db),
		exchangeLogs:     newExchangeLogBucket(exchangeLogBktName, cfg.db),
		unconfirmedTxids: newUnconfirmedTxids(unconfirmedTxsBktName, cfg.db),
		exchgConfig:      cfg,
		quit:             make(chan struct{}),
	}
	// ex.monitor = monitor.New(cfg.depositCoin, ex,
	// 	monitor.Logger(cfg.log),
	// 	monitor.CheckPeriod(cfg.checkPeriod))
	// ex.cli = cli.New(cfg.nodeWltFile, cfg.nodeRPCAddr)
	return ex
}

// Run starts the exchange process
func (ec *exchange) Run() error {
	errC := make(chan error, 1)

	for {
		select {
		case dv := <-ec.scanner.GetDepositValue():
			// TODO: persist the deposit value so that when restart the service,
			// can resend skycoins

			// send skycoins
			// get binded skycoin address
		}
	}

	// eventC := ec.monitor.Run(cxt)

	// go func() {
	// defer ec.log.Debugln("Exchange exit")

	// for {
	// 	select {
	// 	case dv := <-ec.scanner.GetDepositValue():
	// 		if err := ec.eventHandler(e); err != nil {
	// 			errC <- err
	// 			return
	// 		}
	// 	case <-time.After(8 * time.Second):
	// 		var toRemove []struct {
	// 			Txid  string
	// 			Logid int
	// 		}
	// 		if err := ec.unconfirmedTxids.forEach(func(txid string, logid int) {
	// 			// check if this transaction has been executed
	// 			tx, err := ec.cli.GetTransaction(txid)
	// 			if err != nil {
	// 				ec.log.Println("Get transaction failed, err:", err)
	// 				return
	// 			}
	// 			if tx.Transaction.Status.Confirmed {
	// 				ec.log.Printf("Tx %s is executed\n", txid)
	// 				toRemove = append(toRemove, struct {
	// 					Txid  string
	// 					Logid int
	// 				}{txid, logid})
	// 			}
	// 		}); err != nil {
	// 			ec.log.Println(err)
	// 			continue
	// 		}

	// 		for _, item := range toRemove {
	// 			// update the exchange log's tx status
	// 			if err := ec.exchangeLogs.update(item.Logid, func(log *daemon.ExchangeLog) {
	// 				log.Tx.Confirmed = true
	// 			}); err != nil {
	// 				ec.log.Println(err)
	// 			}

	// 			if err := ec.unconfirmedTxids.delete(item.Txid); err != nil {
	// 				ec.log.Println(err)
	// 			}
	// 		}
	// 	}
	// }
	// }()
	return nil
}

func (ec *exchange) AddMonitor(cv coinValue) error {
	if err := ec.coinValue.put(cv); err != nil {
		return err
	}
	return ec.monitor.AddAddress(cv.Address)
}

// GetAllLocalBalances returns all deposit address's balance
func (ec *exchange) GetAllLocalBalances() map[string]int64 {
	v, err := ec.coinValue.getAllBalances()
	if err != nil {
		return map[string]int64{}
	}
	return v
}

func (ec *exchange) GetRealtimeBalance(addr string) (int64, error) {
	url := fmt.Sprintf("https://blockexplorer.com/api/addr/%s/balance", addr)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("Get url:%s fail, error:%s", addr, err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("Read data from resp body fail, error:%s", err)
	}

	defer resp.Body.Close()

	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Parse balance value:%v failed, err:%v", string(data), err)
	}
	return v, nil
}

// GetLogs returns logs in the id range of start and end.
func (ec *exchange) GetLogs(start, end int) ([]daemon.ExchangeLog, error) {
	return ec.exchangeLogs.get(start, end)
}

// GetLogsLen returns the logs length
func (ec *exchange) GetLogsLen() int {
	return ec.exchangeLogs.len()
}

// eventHandler will process the event that was generated by the monitor
func (ec *exchange) eventHandler(e monitor.Event) error {
	if e.Type == monitor.EBalanceChange {
		av, ok := e.Value.(monitor.AddressValue)
		if !ok {
			ec.log.Println("Assert monitor.AddressValue failed")
			return nil
		}

		// send coins
		if av.Value > 0 {
			now := time.Now()
			rate, ok := ec.rateTable.get(now)
			if !ok {
				return fmt.Errorf("Get rate at %v failed", now)
			}

			amount := btcToSkycoinAmount(rate, av.Value)
			cv, ok := ec.coinValue.get(av.Address)
			if !ok {
				ec.log.Printf("Address %s has no ico address\n", av.Address)
				return nil
			}

			ec.log.Println("ico address:", cv.ICOAddress, " amount:", int64(amount))
			txid, err := ec.cli.Send(cv.ICOAddress, int64(amount))
			if err != nil {
				ec.log.Println(err)
				return nil
			}

			ec.log.Debugf("Send %d skycoin to %s, txid:%s", amount, cv.ICOAddress, txid)

			// check the transaction till it's executed or beyond 15 seconds, otherwise returns error
			// if we don't wait untill the transaction is executed, the next sending coin action will
			// failed with error do not have enough balance in wallet.
			var confirmed bool
			for i := 0; i < 7; i++ {
				time.Sleep(2 * time.Second)
				tx, err := ec.cli.GetTransaction(txid)
				if err != nil {
					ec.log.Println(err)
					return nil
				}
				if tx.Transaction.Status.Confirmed {
					confirmed = true
					ec.log.Debugf("Transaction %s is executed\n", txid)
					break
				}
			}

			// records the exchange log
			log := makeExchangeLog(ec.depositCoin, uint64(av.Value), ec.icoCoin, uint64(amount), txid, confirmed)
			if err := ec.exchangeLogs.put(&log); err != nil {
				ec.log.Printf("Record exchange log failed, err :%v", err)
			}

			if !confirmed {
				ec.log.Printf("Transaction %s was not executed in 15 second, maybe master node is crashed.\n", txid)
				// records the unconfirmed transaction
				ec.log.Printf("Records unconfirmed transaction %s", txid)
				if err := ec.unconfirmedTxids.put(txid, log.ID); err != nil {
					ec.log.Printf("Records unconfirmed txid %s with logid %d failed, err:%v\n", txid, log.ID, err)
				}
			}
		}

		return ec.coinValue.changeBalance(av.Address, av.Value)
	}
	// unknow event
	ec.log.Debugf("Unknow event type:%s\n", e.Type)
	return nil
}

func makeExchangeLog(depositAddr string, depositValue uint64, icoAddr string, icoValue uint64, txid string, confirmed bool) daemon.ExchangeLog {
	log := daemon.ExchangeLog{Time: time.Now()}
	log.Deposit.Address = depositAddr
	log.Deposit.Coin = depositValue
	log.ICO.Address = icoAddr
	log.ICO.Coin = icoValue
	log.Tx.Hash = txid
	log.Tx.Confirmed = confirmed
	return log
}

func btcToSkycoinAmount(rate float64, amount int64) int64 {
	return int64(float64(amount) / 1e9 * rate)
}

func (rt *rateTable) loadFromConfig(rates []gconfig.ExchangeRate) {
	_, offset := time.Now().Zone()
	for _, rate := range rates {
		date, err := time.Parse(gconfig.RateTimeLayout, rate.Date)
		if err != nil {
			panic("Invalid exchange rate date")
		}
		date = date.Add(time.Duration(-offset) * time.Second)
		rt.rates = append(rt.rates, struct {
			Date time.Time
			Rate float64
		}{
			date,
			rate.Rate,
		})
	}
}

func (rt *rateTable) get(now time.Time) (float64, bool) {
	// sort the table by time in descending order
	sort.Slice(rt.rates, func(i, j int) bool {
		return rt.rates[i].Date.UnixNano() > rt.rates[j].Date.UnixNano()
	})
	nano := now.UnixNano()
	for _, rate := range rt.rates {
		if nano >= rate.Date.UnixNano() {
			return rate.Rate, true
		}
	}
	return 0.0, false
}
