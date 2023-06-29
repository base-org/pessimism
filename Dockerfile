# Pull golang alpine to build binary
FROM golang:1.19.9-alpine3.16 as builder

RUN apk add --no-cache make gcc musl-dev linux-headers jq bash

WORKDIR /app

COPY . .
RUN make build-app

# Use alpine to run app
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/bin/pessimism .
COPY config.env .
COPY genesis.json .

# Run app and expose api and metrics ports
EXPOSE 8080 7300
CMD ["./pessimism"]
