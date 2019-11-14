export GO111MODULE=on

.PHONY: test lint cover html

GOLINT := $(shell command -v golangci-lint)

test:
	go clean -testcache && go test -race -v .

cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic

html: cover
	go tool cover -html=coverage.txt

lint:
ifndef GOLINT
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
endif
	golangci-lint run -v
