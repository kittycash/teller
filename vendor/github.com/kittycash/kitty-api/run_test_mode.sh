#!/usr/bin/env bash

echo ""
echo "Use the following private key for submitting entries:"
echo "PRIVATE KEY: 190030fed87872ff67015974d4c1432910724d0c0d4bfbd29d3b593dba936155"
echo ""

echo "STARTING 'kitty-api' ..."
KITTY_API_TEST_MODE=true \
KITTY_API_REDIS_ADDRESS=":6379" \
KITTY_API_HTTP_ADDRESS=":7080" \
KITTY_API_RPC_ADDRESS=":7000" \
KITTY_API_MASTER_PUBLIC_KEY="03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753" \
go run ${GOPATH}/src/github.com/kittycash/kitty-api/cmd/kittyapi/kittyapi.go
echo ""