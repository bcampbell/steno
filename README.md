# Steno

Main website at: http://stenoproject.org

This repo contains the client-side Steno tools:

- `steno` - the Steno GUI app to search, tag and analyse collected news articles.
- [`steno-similar`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar) - performs bulk matching between sets of articles
- [`steno-similar-gui`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar-gui) - GUI version of `steno-similar`

### Status

The Steno GUI tool is in quite a bit of flux at the moment.

It originally used the [go-qml](https://github.com/go-qml/qml) bindings to provide a Qt/QML GUI. However, go-qml hasn't really been maintained recently, and packaging up builds with the Qt libraries always got a little awkward.

The current plan is to switch over to a [libui](https://github.com/andlabs/libui)-based GUI.
There is a `ui` branch of steno with a rough proof-of-concept version already working. It's still missing a bunch of features, but is simple, fast and easy to package. it looks promising.


