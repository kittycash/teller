package agent

import (
	"github.com/kittycash/kitty-api/src/rpc"
	"github.com/sirupsen/logrus"
)

type KittyAPIClient struct {
	c *rpc.Client
}

func NewKittyAPI(config *rpc.ClientConfig, log logrus.FieldLogger) *KittyAPIClient {
	client, err := rpc.NewClient(config)
	if err != nil {
		log.Panic(err)
	}

	return &KittyAPIClient{
		c: client,
	}
}
