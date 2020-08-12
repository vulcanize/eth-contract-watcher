FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers

# Get and build eth-contract-watcher
ADD . /go/src/github.com/vulcanize/eth-contract-watcher
WORKDIR /go/src/github.com/vulcanize/eth-contract-watcher
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-contract-watcher .

# Build migration tool
WORKDIR /
RUN go get -d github.com/pressly/goose/cmd/goose
WORKDIR /go/src/github.com/pressly/goose/cmd/goose
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -tags='no_mysql no_sqlite' -o goose .

WORKDIR /go/src/github.com/vulcanize/eth-contract-watcher

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
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/dockerfiles/super_node/startup_script.sh .
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/dockerfiles/super_node/entrypoint.sh .


# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/eth-contract-watcher eth-contract-watcher
COPY --from=builder /go/src/github.com/pressly/goose/cmd/goose/goose goose
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/db/migrations migrations/vulcanizedb
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/environments environments

ENTRYPOINT ["/app/startup_script.sh"]