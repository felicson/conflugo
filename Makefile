ifneq (,$(wildcard ./.env))
    include .env
    export
else
	CONFLUENCE_LOGIN=user
	CONFLUENCE_PASSWORD=passwd
	CONFLUENCE_URL=https://example.com
	CONFLUENCE_SPACE=example
endif

BINARY_NAME=conflugo
fmt:
	@go fmt ./...

run:fmt
	@go run \
		-ldflags '-X main.StorageLogin=$(CONFLUENCE_LOGIN) -X main.StoragePassword=$(CONFLUENCE_PASSWORD) -X main.ConfluenceURL=$(CONFLUENCE_URL) -X main.ConfluenceSpace=$(CONFLUENCE_SPACE)' \
 		cmd/main.go

build:fmt
	@go build -o $(BINARY_NAME) \
		-ldflags '-X main.StorageLogin=$(CONFLUENCE_LOGIN) -X main.StoragePassword=$(CONFLUENCE_PASSWORD) -X main.ConfluenceURL=$(CONFLUENCE_URL) -X main.ConfluenceSpace=$(CONFLUENCE_SPACE)' \
 		cmd/main.go


lint: lint_install ## Lint the source files
	golangci-lint run --timeout 5m

lint_install:
	@which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2

imports:
	goimports -l -w .

imports_install:
	@which goimports || go get -u golang.org/x/tools/cmd/goimports