package main

import (
	"os"
	"path/filepath"

	img64 "github.com/tenkoh/goldmark-img64"
	"github.com/yuin/goldmark"
)

func main() {
	if len(os.Args) != 2 {
		panic("markdown file path is required as an argument")
	}

	mdPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	parentPath := filepath.Base(mdPath)

	b, err := os.ReadFile(mdPath)
	if err != nil {
		panic(err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(img64.Img64),
		goldmark.WithRendererOptions(
			img64.WithPathResolver(
				img64.ParentLocalPathResolver(parentPath),
			),
		),
	)

	md.Convert(b, os.Stdout)
}
