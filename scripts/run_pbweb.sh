#!/bin/bash
source `dirname $0`/env.sh

PBWEB_LDFLAGS="-s -w -X main.version=DEV -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"

go run -ldflags="$PBWEB_LDFLAGS" $MAIN/cmd/pbweb/pbweb.go