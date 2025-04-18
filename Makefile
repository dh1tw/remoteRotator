#!/usr/bin/env bash

SHELL := /bin/bash

PKG := github.com/dh1tw/remoteRotator
COMMITID := $(shell git describe --always --long --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --tags)

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

all: build

build:
	go generate ./...
	go build -v -ldflags="-X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"

# replace the debug version of js libraries with their production, minified versions
js-production:
	find hub/html/index.html -exec sed -i '' 's/vue.js/vue.min.js/g' {} \;

# replace the minified versions of js libraries with their full, development versions
js-development:
	find hub/html/index.html -exec sed -i '' 's/vue.min.js/vue.js/g' {} \;

generate:
	go generate ./...

# strip off dwraf table - used for travis CI
dist:
	go build -v -ldflags="-w -s -X github.com/dh1tw/remoteRotator/cmd.commitHash=${COMMIT} \
		-X github.com/dh1tw/remoteRotator/cmd.version=${VERSION}"
	# compress binary
	# there is a know issue that upx currently doesn't work with darwin/arm64.
	# See https://github.com/upx/upx/issues/424
	# until it's resolved, we ship darwin/arm64 uncompressed.
	if [ "${GOOS}" = "windows" ]; \
		then upx remoteRotator.exe; \
	else \
		if [ "${GOOS}" = "darwin" ]; \
			then true; \
		else upx remoteRotator; \
		fi \
	fi

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

install-deps:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/GeertJohan/go.rice/rice

clean:
	-@rm remoteRotator remoteRotator-v*

.PHONY: build vet lint clean install-deps generate js-production js-development