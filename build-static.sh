#!/bin/bash

# The nsre-alpine-build image is built localy using the Dockerfile.nsre-alpine-build

# docker build -t nsre-alpine-build:latest -f Dockerfile.nsre-alpine-build .

VER=$(git rev-parse --short HEAD)
sed -i "s/const Version = .*/const Version = \"${VER}\"/" cmd/version.go

docker run --rm -v $(pwd):/work --workdir /work --entrypoint go nsre-alpine-build build --tags "icu json1 fts5 secure_delete" --ldflags '-extldflags "-static" -w -s'

mv nsre ~/Public/nsre-static-alpine
