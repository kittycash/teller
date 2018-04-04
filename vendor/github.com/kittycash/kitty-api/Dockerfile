# -*- mode: dockerfile-mode -*-

FROM golang:1.9

WORKDIR /go/src/app/vendor/github.com/kittycash/kitty-api
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENV KITTY_API_REDIS_ADDRESS=:6379
ENV KITTY_API_HTTP_ADDRESS=:7080
ENV KITTY_API_RPC_ADDRESS=:7000
ENV KITTY_API_MASTER_PUBLIC_KEY=03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753
ENV KITTY_API_TEST_MODE=true

ENTRYPOINT ["make"]
CMD ["run"]
