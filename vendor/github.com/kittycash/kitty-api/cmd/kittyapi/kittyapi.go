package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/kittycash/kitty-api/src/api"
	"github.com/kittycash/kitty-api/src/database/redisdb"
	"github.com/kittycash/kitty-api/src/rpc"
	"github.com/kittycash/wallet/src/util"
)

func main() {
	quit := util.CatchInterrupt()
	var (
		testMode      = os.Getenv("KITTY_API_TEST_MODE")
		redisAddress  = os.Getenv("KITTY_API_REDIS_ADDRESS")
		redisPassword = os.Getenv("KITTY_API_REDIS_PASSWORD")
		httpAddress   = os.Getenv("KITTY_API_HTTP_ADDRESS")
		rpcAddress    = os.Getenv("KITTY_API_RPC_ADDRESS")
		masterPubKey  = os.Getenv("KITTY_API_MASTER_PUBLIC_KEY")
	)

	var (
		vTestMode, _ = strconv.ParseBool(testMode)
		dbIndex      = 0
	)

	db, err := redisdb.New(&redisdb.Config{
		Address:  redisAddress,
		TestMode: vTestMode,
		Password: redisPassword,
		Database: dbIndex,
	})
	if err != nil {
		fmt.Printf("Error on db start: %s", err.Error())
		panic(err)
	}
	defer db.Close()

	{
		httpServer := &http.Server{
			Addr:    httpAddress,
			Handler: api.ServePublic(db),
		}
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				fmt.Printf("Error on serving public api: %s", err.Error())
				panic(err)
			}
		}()
		fmt.Printf("Public 'http' api listening on: %s\n", httpServer.Addr)
	}

	{
		rpcServer, err := rpc.NewServer(
			&rpc.ServerConfig{
				Address:   rpcAddress,
				TrustedPK: masterPubKey,
			}, db,
		)
		if err != nil {
			panic(err)
		}
		defer rpcServer.Close()
		fmt.Printf("Private 'rpc' api listening on: %s\n", rpcAddress)
	}

	<-quit
}
