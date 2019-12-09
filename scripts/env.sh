#!/bin/bash
MAIN=`realpath "$(dirname $0)/.."`
BUILD_DIR=$MAIN/builds
LANGUAGE_DIR=$MAIN/language
TEMPLATES_DIR=$MAIN/templates_sample
PB_VERSION=`cat $MAIN/cmd/poundbot/VERSION`
PB_LDFLAGS="-s -w -X main.version=$PB_VERSION -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
