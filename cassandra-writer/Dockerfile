FROM golang:1.24.2-alpine AS builder

WORKDIR /build
COPY cassandra-writer/go.mod cassandra-writer/go.sum cassandra-writer/
COPY cassandra-writer/ cassandra-writer/
COPY common common/
WORKDIR /build/cassandra-writer
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN go build -o cassandra-writer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/cassandra-writer/cassandra-writer .
CMD ["./cassandra-writer"]