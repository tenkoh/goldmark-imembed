# goldmark-img64
## Overview
`goldmark-img64` is a plugin for [goldmark](https://github.com/yuin/goldmark) that automatically embeds image files as Base64-encoded data directly into rendered HTML. This is especially useful for scenarios where:
- Hosting external image files is not practical.
- You need a single self-contained HTML file.

For example, the following Markdown image will be embedded as Base64 data:

```markdown:input.md
![alt](./image.png "title")
```

```html:output.html
<p><img src="data:image/png;base64,..." alt="alt" title="title"></p>
```

This package supports these image types.

- "image/apng"
- "image/avif"
- "image/gif"
- "image/jpeg"
- "image/png"
- "image/svg+xml"
- "image/webp"

## Installation
```sh
go get github.com/tenkoh/goldmark-img64
```
## Limitations
By default, only images that satisfy all the following conditions will be embedded as Base64-encoded data. Other images will have their file paths included as-is in the output.

### Conditions
- **Stored locally**: The image file must reside on the local file system.
- **Valid path**: The image path must be:
  - A relative path based on the executable's directory, or
  - An absolute path.

### Notes
- Remote image URLs (e.g., `https://example.com/image.png`) are not supported by default. See [Embedding Remote Images](#embedding-remote-images) for how to handle such cases.

## Usage recipes
Here are various ways to use `goldmark-img64`:

1. [Basic Usage](#basic-usage)
2. [Handling Non-Standard Paths](#handling-non-standard-paths)
3. [Embedding Remote Images](#embedding-remote-images)
4. [Combining PathResolver and FileReader](#combining-pathresolver-and-filereader)

### Basic Usage
The following example demonstrates how to use `goldmark-img64` with its default settings to convert a Markdown file (`target.md`) into HTML with embedded Base64-encoded images.

```go:example.go
package main

import (
    "io"
    "os"

    "github.com/yuin/goldmark"
    img64 "github.com/tenkoh/goldmark-img64"
)

func main() {
    b, _ := os.ReadFile("target.md")
    goldmark.New(goldmark.WithExtensions(img64.Img64)).Convert(b, os.Stdout)
}
```

### Handling Non-Standard Paths
When your target markdown is not in the current working directory, try `WithPathResolver` option. You can change paths as you like.

`PathResolver` is just a function `func(path string) string`. This repository now provides a simple implementation: `ParentLocalPathResolver`, which adds parent directory's path.

```go
dir := "/path/to/parent"

goldmark.New(
    goldmark.WithExtensions(img64.Img64),
    goldmark.WithRendererOptions(
        img64.WithPathResolver(img64.ParentLocalPathResolver(dir)),
    ),
)
```

### Embedding Remote Images
When you want to embed non local images (ex. `https://..../sample.png`) or want to add pre/post process, you can customize `FileReader`.

`FileReader` is just a function `func(path string) ([]byte, error)`. This repository provides an example: `AllowRemoteFileReader`, which allows embed online images.

```go
goldmark.New(
    goldmark.WithExtensions(img64.Img64),
    goldmark.WithRendererOptions(
        img64.WithFileReader(img64.AllowRemoteFileReader(http.DefaultClient)),
    ),
)
```

### Combining PathResolver and FileReader
You can use both `WithPathResolver` and `WithFileReader` at the same time.

## License
MIT

## Author
tenkoh
