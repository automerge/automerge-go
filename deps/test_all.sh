#!/usr/bin/env bash

set -e
set -x

cd "$(dirname "$(go env GOMOD)")"
GOOS=darwin GOARCH=arm64 go test ./...
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go test
docker run --platform linux/arm64 -v "$(realpath):/automerge" -it --workdir /automerge golang go test
docker run --platform linux/amd64 -v "$(realpath):/automerge" -it --workdir /automerge golang go test