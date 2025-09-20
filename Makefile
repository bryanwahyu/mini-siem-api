APP?=server-analyst
BIN?=$(APP)
PKG=./cmd/server-analyst

.PHONY: build test tidy

build:
	GO111MODULE=on go build -ldflags='-s -w' -o $(BIN) $(PKG)

test:
	go test ./...

tidy:
	go mod tidy
