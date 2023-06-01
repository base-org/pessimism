FROM golang:1.19.0-alpine3.15 as base

## Install dependencies
RUN apk add --no-cache bash ca-certificates git gcc g++ libc-dev make
WORKDIR /go/src/github.com/base-org/pessimism
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
COPY config.env .

## Build binary
FROM base AS builder
COPY . .
RUN make build-app

## Build image
FROM alpine:3.15
RUN apk add ca-certificates
COPY --from=builder /go/src/github.com/base-org/pessimism/bin/pessimism /bin/pessimism
COPY --from=builder /go/src/github.com/base-org/pessimism/config.env .

## Run
CMD ["/bin/pessimism"]