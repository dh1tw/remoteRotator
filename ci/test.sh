#!/bin/bash


if [ "${GIMME_OS}" = "darwin" ] && [ "${GIMME_ARCH}" = "amd64" ]; then
    go test ./...
fi

if [ "${GIMME_OS}" = "linux" ] && [ "${GIMME_ARCH}" = "amd64" ]; then
    $HOME/gopath/bin/goveralls -service=travis-ci
    go test ./...
fi

