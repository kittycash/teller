package main

import (
	"os"
	"os/signal"

	"github.com/skycoin/net/skycoin-messenger/factory"
	"gopkg.in/sirupsen/logrus.v1"
	"gopkg.in/urfave/cli.v1"
)

const (
	AddressFlag = "address"

	DefaultAddress = ":8880"
)

func Flag(flag string, short ...string) string {
	if len(short) == 0 {
		return flag
	}
	return flag + ", " + short[0]
}

var (
	log = logrus.New()
	app = cli.NewApp()
)

func init() {
	log.Out = os.Stdout
	app.Name = "discovery"
	app.Description = "discovery node"
	app.Flags = cli.FlagsByName{
		cli.StringFlag{
			Name:  Flag(AddressFlag, "a"),
			Usage: "address to serve discovery node on",
			Value: DefaultAddress,
		},
	}
	app.Action = cli.ActionFunc(action)
}

func action(ctx *cli.Context) error {
	quit := CatchInterrupt()

	var (
		address = ctx.String(AddressFlag)
	)

	f := factory.NewMessengerFactory()
	f.SetLoggerLevel(factory.InfoLevel)

	if err := f.Listen(address); err != nil {
		return err
	}
	log.Infof("listening on '%s'", address)
	log.Infof("exiting with code '%d'", <-quit)
	return nil
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.WithError(e).Error("exited with error")
	}
}

// CatchInterrupt catches Ctrl+C behaviour.
func CatchInterrupt() chan int {
	quit := make(chan int)
	go func(q chan<- int) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		signal.Stop(sigChan)
		q <- 1
	}(quit)
	return quit
}
