#!/usr/bin/env bash

export GOOS=darwin
export GOARCH

for GOARCH in amd64 arm64; do
    go build -o "./stub-binary-${GOOS}_${GOARCH}" ./internal/stub
done
