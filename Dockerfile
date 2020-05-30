FROM golang:alpine as builder
RUN apk --update --no-cache add make git g++

# Build statically linked vDB binary (wonky path because of Dep)
RUN mkdir -p /go/src/github.com/vulcanize/eth-contract-watcher
ADD . /go/src/github.com/vulcanize/eth-contract-watcher
WORKDIR /go/src/github.com/vulcanize/eth-contract-watcher
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' .

# Build migration tool
RUN go get -u -d github.com/pressly/goose/cmd/goose
WORKDIR /go/src/github.com/pressly/goose/cmd/goose
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -tags='no_mysql no_sqlite' -o goose

# Second stage
FROM alpine
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/eth-contract-watcher /app/eth-contract-watcher
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/environments/staging.toml /app/environments/
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/dockerfiles/startup_script.sh /app/
COPY --from=builder /go/src/github.com/vulcanize/eth-contract-watcher/db/migrations/* /app/
COPY --from=builder /go/src/github.com/pressly/goose/cmd/goose/goose /app/goose

WORKDIR /app
CMD ["./startup_script.sh"]
