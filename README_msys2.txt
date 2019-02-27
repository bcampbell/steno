using the 'offical' Qt dist, the mingw binaries.
However, the pkgconfig .pc files are all screwed up - need to manually hack to produce release versions.


export PKG_CONFIG_PATH="c:/Qt/5.7/mingw53_32/lib/pkgconfig_RELEASE"
#export PKG_CONFIG_PATH="c:/Qt/5.7/mingw53_32/lib/pkgconfig"
export PATH="$PATH:/c/Qt/5.7/mingw53_32/bin"

