#!/bin/bash
set -e

# build script for building a 64bit version of steno under
# msys2, using the msys2 Q1 binaries

STAMP=$(date +"%Y-%m-%d")
OUT=steno_windows_64bit_$STAMP

# where to find the dlls from msys2
MSYSBIN=/c/msys64/mingw64/bin


mkdir $OUT

# TODO:
# sanitycheck presence of bin/fasttext.exe

cp steno/steno.exe $OUT/
cp -r steno/scripts $OUT/
cp -r steno/ui $OUT/
cp -r steno/bin $OUT/
cp steno/slurp_sources.csv $OUT/

# use windeployqt to sort out all the Qt gubbins
#windeployqt --verbose=1 --compiler-runtime --debug --qmldir=$OUT $OUT/steno.exe
windeployqt --verbose=1 --compiler-runtime --release --qmldir=$OUT $OUT/steno.exe

# some stuff is missed by windeployqt:
# we're using the Qt from msys2, so there are a whole bunch of other dlls
# we need to pack up for non-msys2 systems...
#  - see windows_extra_dlls.txt. List just via trial and error, running
#    the exe and seeing what is missing
#
for a in `cat windows_extra_dlls.txt`;
do
    if [ ! -f $OUT/$a ]
    then
        echo "adding $a"
        cp "$MSYSBIN/$a" "$OUT/$a"
    fi
done


