#!/bin/bash
HERE=`dirname $0`
cd $HERE
BUILD_DIR=$HERE/builds
mkdir -p $BUILD_DIR
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR/linux
mkdir -p $BUILD_DIR/darwin
mkdir -p $BUILD_DIR/windows
LDFLAGS="-s -w -X main.Version=`cat $HERE/VERSION` -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
env GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/windows/poundbot.exe poundbot.go
env GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/darwin/poundbot poundbot.go
env GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/linux/poundbot poundbot.go
upx $BUILD_DIR/linux/poundbot