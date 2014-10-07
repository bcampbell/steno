#!/bin/bash
set -e

STAMP=$(date +"%Y-%m-%d")

SCRATCH=$(mktemp -d)
DEST=$SCRATCH/steno

TARBALL=steno_$STAMP.zip

mkdir $DEST

pushd steno >/dev/null
cp -r static templates steno.command $DEST/
GOOS=darwin GOARCH=amd64 go build -o $DEST/steno
GOOS=windows GOARCH=amd64 go build -o $DEST/steno.exe

popd >/dev/null

pushd $SCRATCH >/dev/null
zip -r $TARBALL steno
popd >/dev/null

echo "done (results in $SCRATCH)"

