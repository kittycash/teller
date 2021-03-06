.DEFAULT_GOAL := help
.PHONY: teller test lint check format cover help

PACKAGES = $(shell find ./src -type d -not -path '\./src')

teller: ## Run teller. To add arguments, do 'make ARGS="--foo" teller'.
	go run cmd/teller/teller.go ${ARGS}

test: ## Run tests
	go test ./cmd/... -timeout=1m -cover ${PARALLEL}
	go test ./src/addrs/... -timeout=30s -cover ${PARALLEL}
	go test ./src/config/... -timeout=30s -cover ${PARALLEL}
	go test ./src/exchange/... -timeout=4m -cover ${PARALLEL}
	go test ./src/monitor/... -timeout=30s -cover ${PARALLEL}
	go test ./src/scanner/... -timeout=4m -cover ${PARALLEL} ${MIN_SHUTDOWN_WAIT}
	go test ./src/sender/... -timeout=1m -cover ${PARALLEL}
	go test ./src/teller/... -timeout=30s -cover ${PARALLEL}
	go test ./src/util/... -timeout=30s -cover ${PARALLEL}

test-race: ## Run tests with -race. Note: expected to fail, but look for "DATA RACE" failures specifically
	go test ./cmd/... -timeout=1m -race ${PARALLEL}
	go test ./src/addrs/... -timeout=30s -race ${PARALLEL}
	go test ./src/config/... -timeout=30s -race ${PARALLEL}
	go test ./src/exchange/... -timeout=4m -race ${PARALLEL}
	go test ./src/monitor/... -timeout=30s -race ${PARALLEL}
	go test ./src/scanner/... -timeout=4m -race ${PARALLEL} ${MIN_SHUTDOWN_WAIT}
	go test ./src/sender/... -timeout=1m -race ${PARALLEL}
	go test ./src/teller/... -timeout=30s -race ${PARALLEL}
	go test ./src/util/... -timeout=30s -race ${PARALLEL}
lint: ## Run linters. Use make install-linters first.
	vendorcheck ./...
	gometalinter --deadline=3m -j 2 --disable-all --tests --vendor \
		-E deadcode \
		-E errcheck \
		-E gas \
		-E goconst \
		-E gofmt \
		-E goimports \
		-E golint \
		-E ineffassign \
		-E interfacer \
		-E maligned \
		-E megacheck \
		-E misspell \
		-E nakedret \
		-E structcheck \
		-E unconvert \
		-E unparam \
		-E varcheck \
		-E vet \
		./src/... ./cmd/...

check: lint test ## Run tests and linters

cover: ## Runs tests on ./src/ with HTML code coverage
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	go get -u github.com/alecthomas/gometalinter
	gometalinter --vendored-linters --install

format:  # Formats the code. Must have goimports installed (use make install-linters).
	# This sorts imports by [stdlib, 3rdpart, kittycash/teller]
	goimports -w -local github.com/kittycash/teller ./cmd
	goimports -w -local github.com/kittycash/teller ./src
	# This performs code simplifications
	gofmt -s -w ./cmd
	gofmt -s -w ./src

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
