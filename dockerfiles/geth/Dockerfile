FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

# Get and build vulcanize's geth fork
RUN git clone https://github.com/ethereum/go-ethereum.git /go/src/github.com/ethereum/go-ethereum/
WORKDIR /go/src/github.com/ethereum/go-ethereum
RUN git checkout -b geth tags/v1.9.11
RUN make geth

# app container
FROM alpine

RUN apk add --no-cache ca-certificates
WORKDIR /app

# keep binaries immutable
COPY --from=builder /go/src/github.com/ethereum/go-ethereum/build/bin/geth geth

EXPOSE 8545 8546 8547 30303 30303/udp

ENTRYPOINT ["/app/geth"]