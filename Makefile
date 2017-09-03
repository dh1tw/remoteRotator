#!/usr/bin/env bash

PKG := github.com/dh1tw/remoteRotator
COMMITID := $(shell git describe --always --long --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags)

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
all: build

build:
	go generate ./...
	go build -v -ldflags="-X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"

# strip off dwraf table - used for travis CI

generate:
	go generate ./...

dist: 
	go build -v -ldflags="-w -X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"

# test:
# 	@go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

test:
	go generate ./...
	go test ./... 

install: 
	go generate ./...
	go install -v -ldflags="-w -X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"

install-deps:
	go get ./...

windows:
	go generate ./...
	GOOS=windows GOARCH=386 go get ./...
	GOOS=windows GOARCH=386 go build -v -ldflags="-w -X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"


# static: vet lint
# 	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" ${PKG}

server: build
	./remoteRotator server tcp

clean:
	-@rm remoteRotator remoteRotator-v*

.PHONY: build server install vet lint clean install-deps generate