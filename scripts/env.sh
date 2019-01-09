#!/bin/bash
MAIN=`realpath "$(dirname $0)/.."`
BUILD_DIR=$MAIN/builds
PB_VERSION=`cat $MAIN/cmd/poundbot/VERSION`
PB_LDFLAGS="-s -w -X main.version=$PB_VERSION -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
PBWEB_VERSION=`cat $MAIN/cmd/pbweb/VERSION`
PBWEB_LDFLAGS="-s -w -X main.version=$PBWEB_VERSION -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
PLUGIN_VERSION=`egrep "\[Info\(" $MAIN/rust_plugin/PoundBotConnector.cs | cut -d \" -f 6`
