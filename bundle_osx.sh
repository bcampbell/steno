#!/bin/bash
set -e

OUT=.
APP=$OUT/steno.app
STAMP=$(date +"%Y-%m-%d")
DMG=steno_mac_$STAMP.dmg

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
cp steno/slurp_sources.csv $APP/Contents/Resources/


macdeployqt steno.app -verbose=1 -qmldir=steno.app/Contents/Resources

codesign --deep -s "Developer ID Application" $APP
hdiutil create -srcfolder $APP $DMG

exit 0

# fix up borked framework bundles
for fdir in $APP/Contents/Frameworks/Qt*
do
    f=$(basename $fdir .framework)
    echo fix $f
    pushd $APP/Contents/Frameworks/${f}.framework >/dev/null
    pushd Versions >/dev/null
    ln -s 5 Current
    mv ../Resources 5/
    cp /usr/local/Cellar/qt5/5.3.2/lib/${f}.framework/Contents/Info.plist 5/Resources/
    popd >/dev/null
    ln -s Versions/Current/${f} ${f}
    ln -s Versions/Current/Resources Resources
    popd >/dev/null
done

#codesign --deep -s "Developer ID Application" steno.app

# hdiutil create -srcfolder $APP $DMG

#echo "bundled. now do:"
#echo "scp $DMG ben@roland.mediastandardstrust.org:/srv/vhost/steno.mediastandardstrust.org/web/"

