FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers

# Get and build eth-contract-watcher
WORKDIR /go/src/github.com/vulcanize/eth-contract-watcher
ADD . .
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-contract-watcher .

# Copy migration tool
WORKDIR /
ARG GOOSE_VER="v2.6.0"
ADD https://github.com/pressly/goose/releases/download/${GOOSE_VER}/goose-linux64 ./goose
RUN chmod +x ./goose

# app container
FROM alpine

ARG USER="vdm"
ARG CONFIG_FILE="./environments/example.toml"

RUN adduser -Du 5000 $USER
WORKDIR /app
RUN chown $USER /app
USER $USER

# chown first so dir is writable
COPY --chown=$USER:$USER --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/$CONFIG_FILE config.toml
COPY --chown=$USER:$USER --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/dockerfiles/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/eth-contract-watcher eth-contract-watcher
COPY --from=builder /goose goose
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/db/migrations migrations/vulcanizedb
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/environments environments

ENTRYPOINT ["/app/startup_script.sh"]