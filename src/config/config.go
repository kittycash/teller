// Package config is used to records the service configurations
package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/skycoin/skycoin/src/visor"
)

// Config represents the configuration root
type Config struct {
	// Enable debug logging
	Debug bool `mapstructure:"debug"`
	// Run with gops profiler
	Profile bool `mapstructure:"profile"`
	// Where log is saved
	LogFilename string `mapstructure:"logfile"`
	// Where database is saved, inside the ~/.teller-skycoin data directory
	DBFilename string `mapstructure:"dbfile"`

	// Path of BTC addresses JSON file
	BtcAddresses string `mapstructure:"btc_addresses"`
	// Path of SKY addresses JSON file
	SkyAddresses string `mapstructure:"sky_addresses"`

	Teller Teller `mapstructure:"teller"`

	SkyRPC SkyRPC `mapstructure:"sky_rpc"`
	BtcRPC BtcRPC `mapstructure:"btc_rpc"`

	KittyClientAddr string `mapstructure:"kitty_client_addr"`

	BtcScanner   BtcScanner   `mapstructure:"btc_scanner"`
	SkyScanner   SkyScanner   `mapstructure:"sky_scanner"`
	BoxExchanger BoxExchanger `mapstructure:"box_exchanger"`

	Web Web `mapstructure:"web"`

	AdminPanel AdminPanel `mapstructure:"admin_panel"`

	SecKey string `mapstructure:"secret_key"`

	Dummy Dummy `mapstructure:"dummy"`
}

// Teller config for teller
type Teller struct {
	// Max number of btc addresses a skycoin address can bind
	MaxBoundAddresses int `mapstructure:"max_bound_addrs"`
	// Allow address binding
	BindEnabled bool `mapstructure:"bind_enabled"`
}

// SkyRPC config for Skycoin daemon node RPC
type SkyRPC struct {
	Address string `mapstructure:"address"`
}

// BtcRPC config for btcrpc
type BtcRPC struct {
	Server  string `mapstructure:"server"`
	User    string `mapstructure:"user"`
	Pass    string `mapstructure:"pass"`
	Cert    string `mapstructure:"cert"`
}

// BtcScanner config for BTC scanner
type BtcScanner struct {
	// How often to try to scan for blocks
	ScanPeriod            time.Duration `mapstructure:"scan_period"`
	InitialScanHeight     int64         `mapstructure:"initial_scan_height"`
	ConfirmationsRequired int64         `mapstructure:"confirmations_required"`
	Enabled bool   `mapstructure:"enabled"`
}

// SkyScanner config for SKY Scanner
type SkyScanner struct {
	// How often to try to scan for blocks
	ScanPeriod        time.Duration `mapstructure:"scan_period"`
	InitialScanHeight int64         `mapstructure:"initial_scan_height"`
	Enabled bool   `mapstructure:"enabled"`
}

// BoxExchanger config for box sender
type BoxExchanger struct {
	// Number of decimal places to truncate SKY to
	MaxDecimals int `mapstructure:"max_decimals"`
	// Password of kitty wallet
	GenesisKey string `mapstructure:"genesis_key"`
	// How long to wait before rechecking transaction confirmations
	TxConfirmationCheckWait time.Duration `mapstructure:"tx_confirmation_check_wait"`
	// Allow sending of boxes (deposits will still be received and recorded)
	SendEnabled bool `mapstructure:"send_enabled"`
}

// Validate validates the BoxExchanger config
func (c BoxExchanger) Validate() error {
	if errs := c.validate(); len(errs) != 0 {
		return errs[0]
	}

	return nil
}

func (c BoxExchanger) validate() []error {
	var errs []error

	if c.MaxDecimals < 0 {
		errs = append(errs, errors.New("sky_exchanger.max_decimals can't be negative"))
	}

	if uint64(c.MaxDecimals) > visor.MaxDropletPrecision {
		errs = append(errs, fmt.Errorf("sky_exchanger.max_decimals is larger than visor.MaxDropletPrecision=%d", visor.MaxDropletPrecision))
	}

	return errs
}

// Web config for the teller HTTP interface
type Web struct {
	HTTPAddr         string        `mapstructure:"http_addr"`
	HTTPSAddr        string        `mapstructure:"https_addr"`
	StaticDir        string        `mapstructure:"static_dir"`
	AutoTLSHost      string        `mapstructure:"auto_tls_host"`
	TLSCert          string        `mapstructure:"tls_cert"`
	TLSKey           string        `mapstructure:"tls_key"`
	ThrottleMax      int64         `mapstructure:"throttle_max"` // Maximum number of requests per duration
	ThrottleDuration time.Duration `mapstructure:"throttle_duration"`
	BehindProxy      bool          `mapstructure:"behind_proxy"`
}

// Validate validates Web config
func (c Web) Validate() error {
	if c.HTTPAddr == "" && c.HTTPSAddr == "" {
		return errors.New("at least one of web.http_addr, web.https_addr must be set")
	}

	if c.HTTPSAddr != "" && c.AutoTLSHost == "" && (c.TLSCert == "" || c.TLSKey == "") {
		return errors.New("when using web.https_addr, either web.auto_tls_host or both web.tls_cert and web.tls_key must be set")
	}

	if (c.TLSCert == "" && c.TLSKey != "") || (c.TLSCert != "" && c.TLSKey == "") {
		return errors.New("web.tls_cert and web.tls_key must be set or unset together")
	}

	if c.AutoTLSHost != "" && (c.TLSKey != "" || c.TLSCert != "") {
		return errors.New("either use web.auto_tls_host or both web.tls_key and web.tls_cert")
	}

	if c.HTTPSAddr == "" && (c.AutoTLSHost != "" || c.TLSKey != "" || c.TLSCert != "") {
		return errors.New("web.auto_tls_host or web.tls_key or web.tls_cert is set but web.https_addr is not enabled")
	}

	return nil
}

// AdminPanel config for the admin panel AdminPanel
type AdminPanel struct {
	Host string `mapstructure:"host"`
}

// Dummy config for the fake sender and scanner
type Dummy struct {
	Scanner  bool   `mapstructure:"scanner"`
	Sender   bool   `mapstructure:"sender"`
	HTTPAddr string `mapstructure:"http_addr"`
}

// Redacted returns a copy of the config with sensitive information redacted
func (c Config) Redacted() Config {
	if c.BtcRPC.User != "" {
		c.BtcRPC.User = "<redacted>"
	}

	if c.BtcRPC.Pass != "" {
		c.BtcRPC.Pass = "<redacted>"
	}

	return c
}

// Validate validates the config
func (c Config) Validate() error {
	var errs []string
	oops := func(err string) {
		errs = append(errs, err)
	}

	if c.BtcAddresses == "" {
		oops("btc_addresses missing")
	}
	if _, err := os.Stat(c.BtcAddresses); os.IsNotExist(err) {
		oops("btc_addresses file does not exist")
	}
	if c.SkyAddresses == "" {
		oops("sky_addresses missing")
	}
	if _, err := os.Stat(c.SkyAddresses); os.IsNotExist(err) {
		oops("sky file does not exist")
	}

	if !c.Dummy.Sender {
		if c.SkyRPC.Address == "" {
			oops("sky_rpc.address missing")
		}

		// test if skycoin node rpc service is reachable
		conn, err := net.Dial("tcp", c.SkyRPC.Address)
		if err != nil {
			oops(fmt.Sprintf("sky_rpc.address connect failed: %v", err))
		} else {
			if err := conn.Close(); err != nil {
				log.Printf("Failed to close test connection to sky_rpc.address: %v", err)
			}
		}
	}

	if !c.Dummy.Scanner {
		if c.BtcScanner.Enabled {
			if c.BtcRPC.Server == "" {
				oops("btc_rpc.server missing")
			}

			if c.BtcRPC.User == "" {
				oops("btc_rpc.user missing")
			}
			if c.BtcRPC.Pass == "" {
				oops("btc_rpc.pass missing")
			}
			if c.BtcRPC.Cert == "" {
				oops("btc_rpc.cert missing")
			}

			if _, err := os.Stat(c.BtcRPC.Cert); os.IsNotExist(err) {
				oops("btc_rpc.cert file does not exist")
			}
		}
	}

	if c.BtcScanner.ConfirmationsRequired < 0 {
		oops("btc_scanner.confirmations_required must be >= 0")
	}
	if c.BtcScanner.InitialScanHeight < 0 {
		oops("btc_scanner.initial_scan_height must be >= 0")
	}

	exchangeErrs := c.BoxExchanger.validate()
	for _, err := range exchangeErrs {
		oops(err.Error())
	}
	//
	//if !c.Dummy.Sender {
	//	exchangeErrs := c.BoxExchanger.validateWallet()
	//	for _, err := range exchangeErrs {
	//		oops(err.Error())
	//	}
	//}

	if err := c.Web.Validate(); err != nil {
		oops(err.Error())
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "\n"))
}

func setDefaults() {
	// Top-level args
	viper.SetDefault("profile", false)
	viper.SetDefault("debug", true)
	viper.SetDefault("logfile", "./teller.log")
	viper.SetDefault("dbfile", "teller.db")

	// Teller
	viper.SetDefault("teller.max_bound_btc_addrs", 5)

	// SkyRPC
	viper.SetDefault("sky_rpc.address", "127.0.0.1:6430")
	viper.SetDefault("sky_rpc.enabled", true)

	// BtcRPC
	viper.SetDefault("btc_rpc.server", "127.0.0.1:8334")
	viper.SetDefault("btc_rpc.enabled", true)

	// BtcScanner
	viper.SetDefault("btc_scanner.scan_period", time.Second*20)
	viper.SetDefault("btc_scanner.initial_scan_height", int64(492478))
	viper.SetDefault("btc_scanner.confirmations_required", int64(1))

	// SkyExchanger
	viper.SetDefault("sky_exchanger.tx_confirmation_check_wait", time.Second*5)
	viper.SetDefault("sky_exchanger.max_decimals", 3)
	viper.SetDefault("web.bind_enabled", true)
	viper.SetDefault("web.send_enabled", true)

	// Web
	viper.SetDefault("web.http_addr", "127.0.0.1:7071")
	viper.SetDefault("web.static_dir", "./web/build")
	viper.SetDefault("web.throttle_max", int64(60))
	viper.SetDefault("web.throttle_duration", time.Minute)

	// AdminPanel
	viper.SetDefault("admin_panel.host", "127.0.0.1:7711")

	// DummySender
	viper.SetDefault("dummy.http_addr", "127.0.0.1:4121")
	viper.SetDefault("dummy.scanner", false)
	viper.SetDefault("dummy.sender", false)
}

// Load loads the configuration from "./$configName.*" where "*" is a
// JSON, toml or yaml file (toml preferred).
func Load(configName, appDir string) (Config, error) {
	if strings.HasSuffix(configName, ".toml") {
		configName = configName[:len(configName)-len(".toml")]
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath(appDir)
	viper.AddConfigPath(".")

	setDefaults()

	cfg := Config{}

	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}
