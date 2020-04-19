#!/bin/bash

# The nsre-alpine-build image is built localy using the Dockerfile.nsre-alpine-build

# docker build -t nsre-alpine-build:latest -f Dockerfile.nsre-alpine-build .

VER=$(git rev-parse --short HEAD)
sed -i "s/const Version = .*/const Version = \"${VER}\"/" cmd/version.go

echo "If you change templates while developing remember to run "
echo "go-bindata -fs -nomemcopy -o cmd/bindata.go -pkg cmd  templates/..."
echo "and then commit changes into git."

#docker run --rm -v $(pwd):/work --workdir /work --entrypoint go nsre-alpine-build:latest build --tags "icu json1 fts5 secure_delete" --ldflags '-extldflags "-static" -w -s' -o nsre-linux-amd64-static

CGO_ENABLED=1
GOOS=linux
GOARCH=amd64

cat <<EOF > .gobuild-linux-cgo
CGO_ENABLED=$CGO_ENABLED
GOOS=$GOOS
GOARCH=$GOARCH
EOF

docker rm -f golang-alpine-build || true

WORKSPACE=$(pwd)
if [ -f /.dockerenv ]; then # We start docker within docker
    DOCKER_VOL_OPT="--volumes-from $1"
else
    DOCKER_VOL_OPT="-v ${WORKSPACE}:${WORKSPACE}"
fi
# the image shhould have user jenkins - id 1000 gid 1000 already setup
docker run --name golang-alpine-build $DOCKER_VOL_OPT --user jenkins --workdir $WORKSPACE --entrypoint go --env-file .gobuild-linux-cgo golang-alpine-build:latest build --tags "icu json1 secure_delete" --ldflags '-extldflags "-static" -w -s' -o nsre-${GOOS}-${GOARCH}-static main.go

#mv -f nsre-linux-amd64-static ~/Public/nsre-linux-amd64-static
