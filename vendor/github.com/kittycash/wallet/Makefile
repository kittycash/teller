.DEFAULT_GOAL := help
.PHONY: test lint check install-linters help

test: ## Run tests
	go test -v ./cmd/... -race -timeout=1m -cover
	go test -v ./src/... -race -timeout=1m -cover

lint: ## Run linters. Use make install-linters first.
	vendorcheck ./...
	gometalinter --disable-all -E vet -E goimports -E varcheck --tests --vendor ./...

check: lint test ## Run tests and linters

format: ## Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/kittycash/wallet ./cmd
	goimports -w -local github.com/kittycash/wallet ./src

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	go get -u github.com/alecthomas/gometalinter
	gometalinter --vendored-linters --install

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
