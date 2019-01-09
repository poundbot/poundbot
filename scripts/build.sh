#!/bin/bash
source `dirname $0`/env.sh

rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR/linux
# mkdir -p $BUILD_DIR/darwin
# mkdir -p $BUILD_DIR/windows

# env GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/windows/poundbot.exe $MAIN/cmd/poundbot/poundbot.go
# env GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/darwin/poundbot $MAIN/cmd/poundbot/poundbot.go
env GOOS=linux GOARCH=amd64 go build -ldflags="$PB_LDFLAGS" -o $BUILD_DIR/linux/poundbot $MAIN/cmd/poundbot/poundbot.go
upx $BUILD_DIR/linux/poundbot