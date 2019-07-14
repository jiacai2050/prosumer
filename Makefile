GO111MODULE=on

.PHONY: test lint

GOLINT := $(shell command -v golangci-lint)

test:
	go clean -testcache && go test -v ./...

lint:
ifndef GOLINT
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
endif
	golangci-lint run
