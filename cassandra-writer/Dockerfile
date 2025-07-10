FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
COPY ../common ../common
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cassandra-writer

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/cassandra-writer .

ENTRYPOINT ["./cassandra-writer"]