# Stage 1: Builder
FROM golang:1.25.4-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download
COPY . .

# Build the Go app
RUN go build \
    -ldflags="-s -w -X main.version=$(git describe --tags 2>/dev/null || echo 'dev')" \
    -trimpath \
    -o /app/bin/main ./cmd/api/main.go

# Stage 2: Final Image
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/main .
COPY .env.example .env

RUN chown -R appuser:appgroup /app
USER appuser

ENTRYPOINT ["./main"]
