#!/bin/sh

set -e

export GOOS=linux
export GOARCH=amd64

#-ldflags "-X main.Version=$(git describe)"

DIST_DIR="./dist"
BINARY="nr-fh-nozzle"

echo "> distribution directory: $DIST_DIR"
echo "> binary file: $BINARY"

if [ -d "$DIST_DIR" ]; then
  echo "> cleaning: $DIST_DIR"
  rm -rf "$DIST_DIR"
fi

echo "> running: dep ensure"
dep ensure

echo "> running: go build"
go build -ldflags "-X main.Version=$(git describe)" -o $DIST_DIR/$BINARY

echo "> compressing binary"
tar -czvf $DIST_DIR/$BINARY.tar.gz $DIST_DIR/$BINARY

echo "> $(git describe) ready"
