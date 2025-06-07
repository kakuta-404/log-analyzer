FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/clickhouse-writer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/clickhouse-writer .
CMD ["./clickhouse-writer"]