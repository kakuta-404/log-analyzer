FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/rest-api .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/rest-api .
EXPOSE 8081
CMD ["./rest-api"]