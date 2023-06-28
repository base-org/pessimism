FROM golang:1.19.9-alpine3.16 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers jq bash

WORKDIR /app

COPY . .
RUN make build-app

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/bin/pessimism .
COPY config.env .
COPY genesis.example.json .
CMD ["./pessimism"]
