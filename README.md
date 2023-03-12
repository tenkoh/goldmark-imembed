# goldmark-img64
An extention of [goldmark](https://github.com/yuin/goldmark) to embed local images into a rendered html file as base64 encoded data.

```markdown:input.md
![alt](./image.png "title")
```

```html:output.html
<p><img src="data:image/png;base64,..." alt="alt" title="title"></p>
```

## Installation
```sh
go get github.com/tenkoh/goldmark-img64
```

## Quick usage
When your target markdown file is in the current working directory, just do as below.

```go:example.go
package main

import (
    "io"
    "os"

    "github.com/yuin/goldmark"
    img64 "github.com/tenkoh/goldmark-img64"
)

func main() {
    f, _ := os.Open("target.md")
    defer f.Close()
    b, _ := io.ReadAll(f)
    goldmark.New(goldmark.WithExtensions(img64.Img64)).Convert(b, os.Stdout)
}
```

When your target markdown is not in the current working directory, add `WithParentPath` option.

```go
goldmark.New(
    goldmark.WithExtensions(img64.Img64),
    goldmark.WithRendererOptions(img64.WithParentPath(dir)),
)
```

This package supports these image types.

- "image/apng"
- "image/avif"
- "image/gif"
- "image/jpeg"
- "image/png"
- "image/svg+xml"
- "image/webp"

## License
MIT

## Author
tenkoh
