FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/log-drain .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/log-drain .
CMD ["./log-drain"]