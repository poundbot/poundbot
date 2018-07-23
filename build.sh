#!/bin/bash
HERE=`dirname $0`
mkdir -p $HERE/bin
go build -ldflags="-s -w" -o bin/poundbot poundbot.go
upx bin/poundbot
