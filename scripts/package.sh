#!/bin/bash
source `dirname $0`/env.sh
cd $MAIN

echo "Packaging v${VERSION}, Plugin v${PLUGIN_VERSION}"
if [ ! -d builds ]; then
    echo "Builds not found. Run build.sh first."
    exit 1
fi

for x in windows linux darwin
do
    cd $MAIN/builds/$x
    FILE=PoundBot-`[[ $x = "darwin" ]] && echo "OSX" || echo $x`-${VERSION}.zip
    echo "Creating ${FILE}"
    zip -9r PoundBot-`[[ $x = "darwin" ]] && echo "OSX" || echo $x`-${VERSION}.zip .
done
cd $MAIN/builds
cp $MAIN/rust_plugin/PoundbotConnector.cs .
zip -9 PoundbotConnector-Plugin-${PLUGIN_VERSION}.zip PoundbotConnector.cs
