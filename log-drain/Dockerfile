FROM golang:1.24.2-alpine AS builder

WORKDIR /build
COPY log-drain/go.mod log-drain/go.sum log-drain/
COPY log-drain/main.go log-drain/
COPY common common/
WORKDIR /build/log-drain
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
RUN go build -o log-drain .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/log-drain/log-drain .
CMD ["./log-drain"]