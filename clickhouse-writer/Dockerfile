FROM golang:1.24.2-alpine AS builder

WORKDIR /build
COPY clickhouse-writer/go.mod clickhouse-writer/go.sum clickhouse-writer/
COPY clickhouse-writer/ clickhouse-writer/
COPY common common/
WORKDIR /build/clickhouse-writer
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN go build -o clickhouse-writer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/clickhouse-writer/clickhouse-writer .
CMD ["./clickhouse-writer"]