#!/bin/bash
set -e

STAMP=$(date +"%Y-%m-%d")
OUT=steno_win32_$STAMP

mkdir $OUT

cp steno/steno.exe $OUT/
#RESOURCES="main.qml project.qml Query.qml help.html helper.js"
#for RES in $RESOURCES
#do
#    cp steno/$RES $OUT/
#done
cp -r steno/scripts $OUT/
cp -r steno/ui $OUT/
cp steno/slurp_sources.csv $OUT/


#windeployqt --verbose=1 --compiler-runtime --debug --qmldir=$OUT $OUT/steno.exe
windeployqt --verbose=1 --compiler-runtime --release --qmldir=$OUT $OUT/steno.exe

# some stuff missed by windeployqt:
cp /c/Qt/5.3/mingw482_32/bin/libwinpthread-1.dll $OUT/


echo "done. now copy to:"
echo "ben@roland.mediastandardstrust.org:/srv/vhost/steno.mediastandardstrust.org/web/"

