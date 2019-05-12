#!/bin/bash
source `dirname $0`/env.sh

rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR/linux/language
mkdir -p $BUILD_DIR/windows/language
mkdir -p $BUILD_DIR/linux/templates_sample
mkdir -p $BUILD_DIR/windows/templates_sample

env GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/windows/poundbot-$PB_VERSION.exe $MAIN/cmd/poundbot/poundbot.go
# env GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o $BUILD_DIR/darwin/poundbot $MAIN/cmd/poundbot/poundbot.go
env GOOS=linux GOARCH=amd64 go build -ldflags="$PB_LDFLAGS" -o $BUILD_DIR/linux/poundbot-$PB_VERSION $MAIN/cmd/poundbot/poundbot.go
upx $BUILD_DIR/linux/poundbot-$PB_VERSION
upx $BUILD_DIR/windows/poundbot-$PB_VERSION.exe

cp $LANGUAGE_DIR/active*.toml $BUILD_DIR/linux/language
cp $LANGUAGE_DIR/active*.toml $BUILD_DIR/windows/language
cp $TEMPLATES_DIR/* $BUILD_DIR/linux/templates_sample
cp $TEMPLATES_DIR/* $BUILD_DIR/windows/templates_sample
cd $BUILD_DIR/windows
zip -9r $BUILD_DIR/poundbot-$PB_VERSION.win64.zip *
cd $BUILD_DIR/linux
tar czvf $BUILD_DIR/poundbot-$PB_VERSION.linux-amd64.tar.gz *