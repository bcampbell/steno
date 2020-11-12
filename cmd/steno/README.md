# Steno GUI app

uses the golang Qt bindings at https://github.com/therecipe/qt.

## Building

Rather than plain `go build`, use `qtdeploy` to make sure all the Qt moc
generation magic happens. `qtdeploy` can build natively or in a docker
container. Very handy for cross-compiling.

To do a native build:

```
$ qtdeploy build desktop
```

To cross compile a windows static build (after following the therecipe/qt
instructions for installing an appropriate docker image):
```
$ qtdeploy -docker build windows_64_shared
```
(could also do `windows_64_static`, but had some stability issues...)


## Troubleshooting

### modules vs GOPATH

Bleve has now moved entirely to use golang modules. I don't think you can mix
GOPATH and modules, so the Qt bindings need to be installed as a module too.

The Qt bindings wiki covers using them as modules (and vendoring
the into the project). See the platform-specific directions at:
https://github.com/therecipe/qt/wiki/Installation

### Link errors after cross-platform builds

There is a Qt issue when doing cross-platform builds.
For me, a windows docker build caused a subsequent native linux build to fail with link errors like this:

```
/usr/bin/ld: $WORK/b210/_x005.o: in function `StaticQWindowsVistaStylePluginPluginInstance::StaticQWindowsVistaStylePluginPluginInstance()':
../../vendor/github.com/therecipe/qt/core/core_plugin_import.cpp:5: undefined reference to `qt_static_plugin_QWindowsVistaStylePlugin()'
...
```

It appears to be https://github.com/therecipe/qt/issues/429

Removing the `*_plugin_import*` files seemed to do the trick:
```
$ find ../../vendor/github.com/therecipe/qt -type f -iname "*_plugin_import*" | xargs rm
```

