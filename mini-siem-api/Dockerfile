# syntax=docker/dockerfile:1.7

FROM golang:1.22-alpine AS builder
WORKDIR /app

# Copy go modules first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache build-base
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o mini-siem-api ./cmd/mini-siem-api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates libc6-compat \
    && addgroup -S mini && adduser -S mini -G mini
WORKDIR /home/mini

COPY --from=builder /app/mini-siem-api /usr/local/bin/mini-siem-api

EXPOSE 8085
USER mini

ENTRYPOINT ["/usr/local/bin/mini-siem-api"]
