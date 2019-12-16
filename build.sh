#!/bin/bash

VER=$(git rev-parse --short HEAD)
sed -i "s/const Version = .*/const Version = \"${VER}\"/" cmd/version.go
#go build --tags "icu json1 fts5 secure_delete" -ldflags='-s -w'
go build --tags "icu json1 secure_delete" -ldflags='-s -w'
