#!/bin/bash
DIR=`realpath "$(dirname $0)/.."`
echo $DIR
cd $DIR
goi18n extract -outdir language/
cd language
goi18n merge active.*.toml # translate.*.toml