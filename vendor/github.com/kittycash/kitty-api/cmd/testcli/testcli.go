package main

import (
	"os"

	"github.com/kittycash/kitty-api/src/rand"
	"github.com/kittycash/kitty-api/src/rpc"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/cipher"
	"gopkg.in/urfave/cli.v1"
)

var (
	app = cli.NewApp()
	log = logrus.New()
)

func init() {
	app.Name = "testcli"
	app.Usage = "testcli is for generating kitties for test mode"
	flags := cli.FlagsByName{
		cli.StringFlag{
			Name:  "rpc-address, a",
			Usage: "RPC address.",
			Value: ":7000",
		},
		cli.StringFlag{
			Name:  "secret-key, k",
			Usage: "Secret key to use.",
			Value: "190030fed87872ff67015974d4c1432910724d0c0d4bfbd29d3b593dba936155",
		},
		cli.Uint64Flag{
			Name:  "count, c",
			Usage: "Number of kitties to generate.",
			Value: 100,
		},
	}
	app.Flags = flags
	app.Action = action()
}

func action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		var (
			rpcAddress = ctx.String("rpc-address")
			secretKey  = cipher.MustSecKeyFromHex(ctx.String("secret-key"))
			count      = ctx.Uint64("count")
		)
		client, err := rpc.NewClient(&rpc.ClientConfig{
			Address: rpcAddress,
		})
		if err != nil {
			return err
		}
		defer client.Close()

		return client.AddEntries(&rpc.AddEntriesIn{
			Entries: rand.GenerateKitties(count, secretKey),
		})
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
