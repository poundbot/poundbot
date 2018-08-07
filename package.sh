#!/bin/bash
HERE=`dirname \`readlink -f $0\``
cd $HERE
VERSION=`cat VERSION`
PLUGIN_VERSION=`egrep "\[Info\(" rust_plugin/PoundbotConnector.cs | cut -d \" -f 6`
echo "Packaging v${VERSION}, Plugin v${PLUGIN_VERSION}"
if [ ! -d builds ]; then
    echo "Builds not found. Run build.sh first."
    exit 1
fi

for x in windows linux darwin
do
    cd $HERE/builds/$x
    FILE=PoundBot-`[[ $x = "darwin" ]] && echo "OSX" || echo $x`-${VERSION}.zip
    echo "Creating ${FILE}"
    zip -9r PoundBot-`[[ $x = "darwin" ]] && echo "OSX" || echo $x`-${VERSION}.zip .
done
cd $HERE/builds
cp $HERE/rust_plugin/PoundbotConnector.cs .
zip -9 PoundbotConnector-Plugin-${PLUGIN_VERSION}.zip PoundbotConnector.cs
