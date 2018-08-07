MAIN=`go env GOPATH`/src/bitbucket.org/mrpoundsign/poundbot
BUILD_DIR=$MAIN/builds
LDFLAGS="-s -w -X main.Version=`cat $MAIN/VERSION` -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
VERSION=`cat $MAIN/VERSION`
PLUGIN_VERSION=`egrep "\[Info\(" $MAIN/rust_plugin/PoundbotConnector.cs | cut -d \" -f 6`