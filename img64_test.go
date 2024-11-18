package img64

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

const minimizedPng = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVQIHWP4DwABAQEANl9ngAAAAABJRU5ErkJggg=="

func decodeTempImage(t *testing.T, w io.Writer, data string) {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		t.Fatal(err)
	}
	w.Write(b)
}

func TestEncodeImage(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	decodeTempImage(t, f, minimizedPng)
	f.Close()

	txtPath := filepath.Join(dir, "test.txt")
	f, err = os.Create(txtPath)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("Gopher"))
	f.Close()

	tURL := "https://example.com/test.png"
	wData := fmt.Sprintf("data:image/png;base64,%s", minimizedPng)

	tests := []struct {
		name          string
		source        string
		optParentPath string
		want          string
		wantError     bool
	}{
		{"read local image file", imgPath, "", wData, false},
		{"read local non-image file", txtPath, "", "", true},
		{"read online image", tURL, "", tURL, false},
		{"read dataURL", wData, "", wData, false},
		{"can not resolve file path", filepath.Base(imgPath), "", "", true},
		{"can resolve file path", filepath.Base(imgPath), dir, wData, false},
		{"other online resource", "s3://foobucket/bar.png", "", "", true},
	}

	r := &img64Renderer{
		Img64Config: Img64Config{
			Config:     html.NewConfig(),
			FileReader: defaultFileReader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.ParentPath = tt.optParentPath
			got, err := r.encodeImage([]byte(tt.source))
			hasError := err != nil
			if hasError != tt.wantError {
				t.Errorf("unintended error situation. want %t, got %t (%s)", tt.wantError, hasError, err)
				return
			}
			if got := string(got); got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func TestImg64(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	decodeTempImage(t, f, minimizedPng)
	f.Close()

	srcAbs := fmt.Sprintf(`![test](%s "title")`, imgPath)
	srcRel := fmt.Sprintf(`![test](%s "title")`, filepath.Base(imgPath))

	var buf bytes.Buffer

	goldmark.New(goldmark.WithExtensions(Img64)).Convert([]byte(srcAbs), &buf)
	// NOTE: the output includes trailing newline. so do not join the below two lines.
	want := fmt.Sprintf(`<p><img src="data:image/png;base64,%s" alt="test" title="title"></p>
`, minimizedPng)
	if got := buf.String(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}

	buf.Reset()

	goldmark.New(
		goldmark.WithExtensions(Img64),
		goldmark.WithRendererOptions(WithParentPath(dir)),
	).Convert([]byte(srcRel), &buf)
	if got := buf.String(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
