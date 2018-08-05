#!/bin/bash
HERE=`dirname \`readlink -f $0\``
cd $HERE
if [ ! -d builds ]; then
    echo "Builds not found. Run build.sh first."
    exit 1
fi

for x in windows linux darwin
do
    cd $HERE/builds/$x
    zip -9r PoundBot-`[[ $x = "darwin" ]] && echo "OSX" || echo $x`-`cat $HERE/VERSION`.zip .
done