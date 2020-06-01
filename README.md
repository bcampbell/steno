# Steno

Main website at: http://stenoproject.org

This repo contains the client-side Steno tools:

- `cmd/steno` - the Steno GUI app to search, tag and analyse collected news articles.
- [`cmd/steno-similar`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar) - performs bulk matching between sets of articles
- [`cmd/steno-similar-gui`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar-gui) - GUI version of `steno-similar`

## Status

The Steno GUI tool is in a bit of flux at the moment.

It originally used the [go-qml](https://github.com/go-qml/qml) bindings to provide a Qt/QML GUI.
However, go-qml hasn't really been maintained recently, and there are a whole bunch of forks, all in various states of workingness.
So it's being switched to use https://github.com/therecipe/qt.

## Building

TODO: flesh this out

### requirements

- golang
- http://github.com/therecipe/qt

### build

```
$ cd cmd/steno
$ qtdeploy build desktop
```



