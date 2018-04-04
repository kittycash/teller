# Kittycash Wallet

Where one claims ownership of dem' kitties.

## Test wallet

**Start `kitty-api`**

```bash
# Go into directory.
cd ${GOPATH}/src/github.com/kittycash/kitty-api

# Run with environment variable declared.
bash run_test_mode.sh

# Inject some kitties.
go run ${GOPATH}/src/github.com/kittycash/kitty-api/cmd/testcli/testcli.go
```

Detailed Instructions: https://github.com/kittycash/kitty-api

**Start Discovery Node**

```bash
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/discovery/discovery.go \
--address=":8880"
```

**Start IKO Node in Test Mode**

```bash
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/iko/iko.go \
--root-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--root-sk=190030fed87872ff67015974d4c1432910724d0c0d4bfbd29d3b593dba936155 \
--root-nc=1234 \
--tx-gen-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--tx-tran-pks=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--init=true \
--test=true \
--test-tx-gen-count=100 \
--test-tx-gen-sk=190030fed87872ff67015974d4c1432910724d0c0d4bfbd29d3b593dba936155 \
--cxo-address="127.0.0.1:7140" \
--messenger-addresses=":8880" 
```

**Start Wallet Node in Test Mode**

```bash
go run ${GOPATH}/src/github.com/kittycash/wallet/cmd/wallet/wallet.go \
--test=true \
--test-gen-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--test-root-pk=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
--test-root-nonce=1234 \
--test-api-domain="127.0.0.1:7080" \
--cxo-address="127.0.0.1:6140" \
--http-address="127.0.0.1:6148" \
--messenger-addresses=":8880"
```