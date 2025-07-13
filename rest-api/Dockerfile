FROM golang:1.24.2-alpine AS builder

WORKDIR /build
COPY rest-api/go.mod rest-api/go.sum rest-api/
COPY rest-api/ rest-api/
COPY common common/
WORKDIR /build/rest-api
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN go build -o rest-api .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/rest-api/rest-api .
CMD ["./rest-api"]