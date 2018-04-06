# teller build binaries
# reference https://github.com/kittycash/teller
FROM golang:1.9-alpine AS build-go

RUN apk add --no-cache gcc musl-dev linux-headers

COPY . $GOPATH/src/github.com/kittycash/teller

RUN cd $GOPATH/src/github.com/kittycash/teller && \
  CGO_ENABLED=1 GOOS=linux go install -ldflags "-s" -installsuffix cgo ./cmd/...

# teller image
FROM alpine:3.7

ENV DATA_DIR="/data"

RUN adduser -D teller

USER teller

# copy binaries
COPY --from=build-go /go/bin/* /usr/bin/

# copy config
COPY ./config.toml /usr/local/teller/
COPY ./example_btc_addresses.json /usr/local/teller/
COPY ./example_eth_addresses.json /usr/local/teller/

# volumes
VOLUME $DATA_DIR

EXPOSE 4121 7071 7711
WORKDIR /usr/local/teller

CMD ["teller"]
