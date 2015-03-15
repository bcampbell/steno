#!/bin/bash
set -e

OUT=.
APP=$OUT/steno.app
rm -rf $APP
rm -f steno.dmg

mkdir $APP
mkdir $APP/Contents
mkdir $APP/Contents/MacOS
mkdir $APP/Contents/Resources

cp Info.plist $APP/Contents/
cp steno/steno $APP/Contents/MacOS/
cp -r steno/ui $APP/Contents/Resources/
cp -r steno/scripts $APP/Contents/Resources/


/usr/local/Cellar/qt5/5.3.2/bin/macdeployqt steno.app -verbose=1 -qmldir=steno.app/Contents/Resources -dmg


echo "bundled. now do:"
echo "scp steno.dmg ben@roland.mediastandardstrust.org:/srv/vhost/steno.mediastandardstrust.org/web/"

