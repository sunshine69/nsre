#!/bin/sh

VER=$(git rev-parse --short HEAD)
sed -i "s/const Version = .*/const Version = \"${VER}\"/" cmd/version.go

echo "If you change templates while developing remember to run "
echo "go-bindata -fs -nomemcopy -o cmd/bindata.go -pkg cmd  templates/..."
echo "and then commit changes into git."

export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

#go build --tags "icu json1 fts5 secure_delete" -ldflags='-s -w'
#go build --tags "icu json1 secure_delete" -ldflags='-s -w'
go build --tags "icu json1 secure_delete" --ldflags '-extldflags "-static" -w -s' -o nsre-${GOOS}-${GOARCH}-static main.go
