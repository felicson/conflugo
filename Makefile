ifneq (,$(wildcard ./.env))
    include .env
    export
else
	STORAGE_LOGIN=user
	STORAGE_PASSWORD=passwd
endif

BINARY_NAME=conflugo
fmt:
	@go fmt ./...

run:fmt
	@go run -ldflags '-X main.StorageLogin=$(STORAGE_LOGIN) -X main.StoragePassword=$(STORAGE_PASSWORD)' cmd/main.go

build:fmt
	@go build -o $(BINARY_NAME) -ldflags '-X main.StorageLogin=$(STORAGE_LOGIN) -X main.StoragePassword=$(STORAGE_PASSWORD)' cmd/main.go


lint: lint_install ## Lint the source files
	golangci-lint run --timeout 5m

lint_install:
	@which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.37.1

imports:
	goimports -l -w .

imports_install:
	@which goimports || go get -u golang.org/x/tools/cmd/goimports