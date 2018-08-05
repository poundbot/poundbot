#!/bin/bash
HERE=`dirname $0`
cd $HERE
BUILD_DIR=$HERE/builds
mkdir -p $BUILD_DIR
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR/linux
mkdir -p $BUILD_DIR/darwin
mkdir -p $BUILD_DIR/windows
env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/windows/poundbot.exe poundbot.go
env GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/dawrin/poundbot poundbot.go
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/linux/poundbot poundbot.go
upx $BUILD_DIR/linux/poundbot