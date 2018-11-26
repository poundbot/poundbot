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
    cd $BUILD_DIR/$x
    FILE="${BUILD_DIR}/PoundBot-$x-${VERSION}.zip"
    echo "Creating ${FILE}"
    zip -9r $FILE .
done
cd $BUILD_DIR
cp $MAIN/rust_plugin/PoundBotConnector.cs .
zip -9 PoundbotConnector-Plugin-${PLUGIN_VERSION}.zip PoundBotConnector.cs
