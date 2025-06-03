FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/gui .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gui .
EXPOSE 8083
CMD ["./gui"]