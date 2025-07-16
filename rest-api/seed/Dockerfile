# seed/Dockerfile
FROM golang:1.24-alpine

WORKDIR /app
COPY . .
RUN go mod init seed && go mod tidy

CMD ["go", "run", "main.go"]
