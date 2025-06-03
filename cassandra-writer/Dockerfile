FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/cassandra-writer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cassandra-writer .
CMD ["./cassandra-writer"]
