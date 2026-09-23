FROM golang:1.26-alpine AS builder
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN swag init -g cmd/api/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]