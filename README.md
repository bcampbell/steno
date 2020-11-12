# Steno

Main website at: http://stenoproject.org

This repo contains the client-side Steno tools:

- `cmd/steno` - the Steno GUI app to search, tag and analyse collected news articles.
- [`cmd/steno-similar`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar) - performs bulk matching between sets of articles
- [`cmd/steno-similar-gui`](https://github.com/bcampbell/steno/tree/master/cmd/steno-similar-gui) - GUI version of `steno-similar`

### Requirements

- golang
- http://github.com/therecipe/qt

### Build

```
$ cd cmd/steno
$ qtdeploy build desktop
```

See [cmd/steno/README.md](cmd/steno/README.md) for more details.

