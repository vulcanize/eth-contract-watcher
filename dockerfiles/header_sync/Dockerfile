FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers

# Add this repo for the startup_script
ADD . /go/src/github.com/vulcanize/eth-contract-watcher

# Get and build eth-header-sync
RUN git clone https://github.com/vulcanize/eth-header-sync.git /go/src/github.com/vulcanize/eth-header-sync
WORKDIR /go/src/github.com/vulcanize/eth-header-sync
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-header-sync .

# Build migration tool
WORKDIR /
RUN git clone https://github.com/pressly/goose.git /go/src/github.com/pressly/goose/
WORKDIR /go/src/github.com/pressly/goose/cmd/goose
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -tags='no_mysql no_sqlite' -o goose .

WORKDIR /go/src/github.com/vulcanize/eth-header-sync

# app container
FROM alpine

ARG USER
ARG CONFIG_FILE

RUN adduser -Du 5000 $USER
WORKDIR /app
RUN chown $USER /app
USER $USER

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/dockerfiles/header_sync/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/eth-header-sync eth-header-sync
COPY --from=builder /go/src/github.com/pressly/goose/cmd/goose/goose goose
COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/db/migrations migrations/vulcanizedb
COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/environments environments

ENTRYPOINT ["/app/startup_script.sh"]