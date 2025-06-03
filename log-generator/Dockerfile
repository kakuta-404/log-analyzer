FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/log-generator ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/log-generator .
EXPOSE 8080
CMD ["./log-generator"]
