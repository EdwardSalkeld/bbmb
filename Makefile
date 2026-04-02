.PHONY: help build test clean \
	server-build server-build-portable server-build-linux-amd64 server-build-linux-arm64 server-run \
	server-test server-coverage server-fmt-check server-vet \
	go-client-build go-client-test go-client-fmt-check go-client-vet \
	python-client-test

SERVER_DIR := server
GO_CLIENT_DIR := go-client
PY_CLIENT_DIR := python-client

SERVER_BIN := $(SERVER_DIR)/bbmb-server
GO_CLIENT_BIN := $(GO_CLIENT_DIR)/bbmb-client
SERVER_GOFLAGS := -trimpath
SERVER_LDFLAGS := -s -w

help:
	@echo "Available targets:"
	@echo "  server-build                Build server for current OS/arch"
	@echo "  server-build-portable       Build server with CGO disabled (more portable)"
	@echo "  server-build-linux-amd64    Cross-build server for Linux x86_64"
	@echo "  server-build-linux-arm64    Cross-build server for Linux ARM64"
	@echo "  server-run                  Run server"
	@echo "  server-test                 Run server tests"
	@echo "  server-coverage             Run server coverage tests"
	@echo "  go-client-build             Build Go client binary"
	@echo "  go-client-test              Run Go client tests"
	@echo "  python-client-test          Run Python client tests"
	@echo "  build                       Build server and Go client"
	@echo "  test                        Run all tests"
	@echo "  clean                       Remove build artifacts"

server-build:
	cd $(SERVER_DIR) && go build $(SERVER_GOFLAGS) -ldflags='$(SERVER_LDFLAGS)' -o bbmb-server .

server-build-portable:
	cd $(SERVER_DIR) && CGO_ENABLED=0 go build $(SERVER_GOFLAGS) -ldflags='$(SERVER_LDFLAGS)' -o bbmb-server .

server-build-linux-amd64:
	cd $(SERVER_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(SERVER_GOFLAGS) -ldflags='$(SERVER_LDFLAGS)' -o bbmb-server-linux-amd64 .

server-build-linux-arm64:
	cd $(SERVER_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(SERVER_GOFLAGS) -ldflags='$(SERVER_LDFLAGS)' -o bbmb-server-linux-arm64 .

server-run:
	cd $(SERVER_DIR) && ./bbmb-server

server-test:
	cd $(SERVER_DIR) && go test -v ./...

server-coverage:
	cd $(SERVER_DIR) && go test -v -coverprofile=coverage.out ./protocol ./queue ./metrics

server-fmt-check:
	cd $(SERVER_DIR) && test -z "$$(gofmt -s -l .)"

server-vet:
	cd $(SERVER_DIR) && go vet ./...

go-client-build:
	cd $(GO_CLIENT_DIR) && go build -o bbmb-client ./cmd

go-client-test:
	cd $(GO_CLIENT_DIR) && go test -v ./...

go-client-fmt-check:
	cd $(GO_CLIENT_DIR) && test -z "$$(gofmt -s -l .)"

go-client-vet:
	cd $(GO_CLIENT_DIR) && go vet ./...

python-client-test:
	cd $(PY_CLIENT_DIR) && python -m unittest discover -s tests -p "test_*.py"

build: server-build go-client-build

test: server-test go-client-test python-client-test

clean:
	rm -f $(SERVER_BIN) \
		$(SERVER_DIR)/bbmb-server-linux-amd64 \
		$(SERVER_DIR)/bbmb-server-linux-arm64 \
		$(SERVER_DIR)/coverage.out \
		$(GO_CLIENT_BIN)
