package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	img64 "github.com/tenkoh/goldmark-img64"
	"github.com/yuin/goldmark"
)

func remotePathResolver(root url.URL) img64.PathResolver {
	return func(s string) string {
		root.Path = path.Join(root.Path, s)
		return root.String()
	}
}

func main() {
	if len(os.Args) != 2 {
		panic("remote markdown filepath must be specified")
	}

	mdPath := os.Args[1]
	root, err := url.Parse(mdPath)
	if err != nil {
		panic(err)
	}
	root.Path = path.Dir(path.Dir(root.Path)) // this is just a sample. it depends on each case.

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(mdPath)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(img64.Img64),
		goldmark.WithRendererOptions(
			img64.WithPathResolver(remotePathResolver(*root)),
			img64.WithFileReader(
				img64.AllowRemoteFileReader(client),
			),
		),
	)

	md.Convert(b, os.Stdout)
}
