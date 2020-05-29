# Steno GUI app

uses the golang Qt bindings at https://github.com/therecipe/qt.

## Building

Rather than plain `go build`, use qtdeploy to make sure all the Qt moc
generation magic happens:

```
$ qtdeploy build desktop
```


