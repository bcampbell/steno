# Steno

Main website at: http://stenoproject.org

This repo contains the client-side Steno tools:

- `steno` - the Steno GUI app to search, tag and analyse collected news articles.
- [`steno-similar`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar) - performs bulk matching between sets of articles
- [`steno-similar-gui`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar-gui) - GUI version of `steno-similar`

## Status

The Steno GUI tool is in quite a bit of flux at the moment.

It originally used the [go-qml](https://github.com/go-qml/qml) bindings to provide a Qt/QML GUI. However, go-qml hasn't really been maintained recently, and there are a whole bunch of forks, all in various states of workingness.

I've settled upon the https://github.com/jamalsa/qml fork for now.


## Building

(assuming your GOPATH is (`$HOME/go`):

    go get https://github.com/bcampbell/steno
    cd ~/go/src/github.com/bcampbell/steno/steno
    go build

Usually there'll be a whole bunch of packages you need to `go get`.
I'm sure there's a more elegant way to install them, but for now, just
keep running `go build`, followed by `go get .......` to install whatever it complains about, until it builds.

### requirements

A working [Go](https://golang.org) installation is required. Versions >=1.12 won't work with steno currently!

- Qt development files (including private headers)
- a bunch of go packages

The Qt/QML binding used by Steno requires the Qt5 headers, and also the Qt5 private headers. Some Linux distros package the 

    $ sudo apt install qtdeclarative5-dev qtbase5-private-dev qtdeclarative5-private-dev

### Troubleshooting

#### missing private qt headers


    #include <private/qmetaobject_p.h>
    In file included from ./cpp/private/qmetaobject_p.h:1,
                     from cpp/govalue.h:7,
                     from cpp/capi.cpp:10,
                     from all.cpp:2:
    ./cpp/private/qtheader.h:70:37: fatal error: QtCore/5.11.3/QtCore/private/qmetaobject_p.h: No such file or directory


sudo apt install qtbase5-private-dev qtdeclarative5-private-dev



edit bridge.go to add the path to the private includes:

    // #cgo CPPFLAGS: -I./cpp

becomes:

    // #cgo CPPFLAGS: -I./cpp -I/usr/include/x86_64-linux-gnu/qt5/QtCore/5.11.3

or?:

    export CGO_CPPFLAGS="-I/usr/include/x86_64-linux-gnu/qt5/QtCore/5.11.3"


#### Doesn't work on go1.12+

Go changed it ABI in 1.12 in a way which breaks the QML bindings.

Symptom:

    github.com/jamalsa/qml/cdata.Ref: relocation target runtime.acquirem not defined for ABI0 (but is defined for ABIInternal)
    github.com/jamalsa/qml/cdata.Ref: relocation target runtime.releasem not defined for ABI0 (but is defined for ABIInternal)

Solution: use an earlier version of go. (I'm using go 1.10.3)

#### Missing modules at runtime:

    Error starting App: file:////home/ben/go/src/github.com/bcampbell/steno/steno/ui/main.qml.:6 module "QtQuick.Dialogs" is not installed

eg, under debian:

    sudo apt-get install qml-module-qtquick-controls qml-module-qtquick-dialogs qml-module-qtqml-models2 qml-module-qt-labs-folderlistmodel qml-module-qt-labs-settings

## Future Plans

### GUI

Packaging up builds with Qt/QML stuff always got a little awkward - loads of dependencies to include.

The current plan is to switch over to a [libui](https://github.com/andlabs/libui)-based GUI.
There is a `ui` branch of steno with a rough proof-of-concept version already working.
It's still missing a bunch of features, but is simple, fast and easy to package. It looks promising.


