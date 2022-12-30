# Aseprite image loader

[![GoDoc](https://godoc.org/github.com/askeladdk/aseprite?status.png)](https://godoc.org/github.com/askeladdk/aseprite)
[![Go Report Card](https://goreportcard.com/badge/github.com/askeladdk/aseprite)](https://goreportcard.com/report/github.com/askeladdk/aseprite)

## Overview

Package aseprite implements a decoder for Aseprite sprite files (`.ase` and `.aseprite` files).

Layers are flattened, blending modes are applied, and frames are arranged on a single texture atlas. Invisible and reference layers are ignored. Tilesets and external files are not supported.

## Install

```
go get -u github.com/askeladdk/aseprite
```

## Quickstart

Use `image.Decode` to decode an aseprite sprite file to an `image.Image`:

```go
img, imgformat, err := image.Decode("test.aseprite")
```

This works fine when loading single images, but if the sprite contains multiple frames they will be organized as a texture atlas. In that case, type cast the image to a `aseprite.Aseprite` to access the metadata:

```go
if imgformat == "aseprite" {
    sprite := img.(*aseprite.Aseprite)
    for _, frame := range sprite.Frames {
        // etc ...
    }
}
```

## License

Package aseprite is released under the terms of the ISC license.
