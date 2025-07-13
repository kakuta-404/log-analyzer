FROM golang:1.24.2-alpine AS builder

WORKDIR /build
COPY log-generator/go.mod log-generator/go.sum log-generator/
COPY log-generator/ log-generator/
COPY common common/
WORKDIR /build/log-generator
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN go build -o log-generator .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/log-generator/log-generator .
CMD ["./log-generator"]
