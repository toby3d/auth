SHELL:=/bin/bash
PROJECT_NAME=indieauth
VERSION=$(shell cat VERSION)
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_FILES=$(shell go list ./... | grep -v /vendor/)

build:
	$(GO_BUILD_ENV) go build .

run:
	go run .

clean:
	go clean

test:
	go test -race -cover ./...

dep:
	go mod download

gen:
	go generate ./...

lint:
	golangci-lint run

.PHONY: build run clean test dep gen lint