MAIN=`go env GOPATH`/src/bitbucket.org/mrpoundsign/poundbot
BUILD_DIR=$MAIN/builds
VERSION=`cat $MAIN/VERSION`
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`"
PLUGIN_VERSION=`egrep "\[Info\(" $MAIN/rust_plugin/PoundBotConnector.cs | cut -d \" -f 6`
