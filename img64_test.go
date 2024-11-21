package img64

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
)

const (
	minimizedPng     = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVQIHWP4DwABAQEANl9ngAAAAABJRU5ErkJggg=="
	imageFileName    = "image.png"
	nonImageFileName = "test.txt"
)

func writeImageDataFromBase64Data(t *testing.T, w io.Writer, data string) {
	t.Helper()
	d := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	_, err := io.Copy(w, d)
	if err != nil {
		t.Fatal(err)
	}
}

func prepareImageFile(t *testing.T, dir string) (filePath string) {
	t.Helper()
	filePath = filepath.Join(dir, imageFileName)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	writeImageDataFromBase64Data(t, f, minimizedPng)
	return
}

func prepareNonImageFile(t *testing.T, dir string) (filePath string) {
	t.Helper()
	filePath = filepath.Join(dir, nonImageFileName)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fmt.Fprint(f, "Gopher")
	return
}

func TestEncodeImage(t *testing.T) {
	dir := t.TempDir()
	imagePath := prepareImageFile(t, dir)
	nonImagePath := prepareNonImageFile(t, dir)
	wantData := fmt.Sprintf("data:image/png;base64,%s", minimizedPng)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeImageDataFromBase64Data(t, w, minimizedPng)
	}))
	defer testServer.Close()

	tests := []struct {
		name      string
		source    string
		options   []Img64Option
		wantError bool
		want      string
	}{
		{"default: read local image file", imagePath, nil, false, wantData},
		{"default: raise error while reading local non-image file", nonImagePath, nil, true, ""},
		{"default: pass through URL", testServer.URL, nil, false, testServer.URL},
		{"default: pass through dataURL", wantData, nil, false, wantData},
		{"default: raise error if path was not resolved", filepath.Base(imagePath), nil, true, ""},
		{
			"with optional PathResolver: path was resolved",
			filepath.Base(imagePath),
			[]Img64Option{
				WithPathResolver(ParentLocalPathResolver(dir)),
			},
			false,
			wantData,
		},
		{
			"with optional FileReader: remote images was loaded",
			testServer.URL,
			[]Img64Option{
				WithFileReader(AllowRemoteFileReader(http.DefaultClient)),
			},
			false,
			wantData,
		},
		{
			"both FileReader and PathResolver options can be used together",
			testServer.URL,
			[]Img64Option{
				WithFileReader(AllowRemoteFileReader(http.DefaultClient)),
				WithPathResolver(ParentLocalPathResolver("/tmp")),
			},
			false,
			wantData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewImg64Renderer(tt.options...)
			got, err := r.encodeImage([]byte(tt.source))
			gotError := err != nil
			if gotError != tt.wantError {
				t.Errorf("unintended error situation. want %t, got %t (%s)", tt.wantError, gotError, err.Error())
				return
			}
			if err != nil {
				return
			}
			if got := string(got); got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	dir := t.TempDir()
	imagePath := prepareImageFile(t, dir)

	imageElementWithAbsolutePath := fmt.Sprintf(`![test](%s "title")`, imagePath)
	imageElementWithRelativePath := fmt.Sprintf(`![test](%s "title")`, filepath.Base(imagePath))

	// NOTE: the output includes trailing newline. so do not join the below two lines.
	encoded := fmt.Sprintf(`<p><img src="data:image/png;base64,%s" alt="test" title="title"></p>
`, minimizedPng)
	notEncoded := fmt.Sprintf(`<p><img src="%s" alt="test" title="title"></p>
`, filepath.Base(imagePath))

	tests := []struct {
		name      string
		mdText    string
		options   []renderer.Option
		wantError bool
		want      string
	}{
		{
			"successfully encoded without options",
			imageElementWithAbsolutePath,
			nil,
			false,
			encoded,
		},
		{
			"encoding failed without options, output the source path itself",
			imageElementWithRelativePath,
			nil,
			false,
			notEncoded,
		},
		{
			"succeeded to use a option",
			imageElementWithRelativePath,
			[]renderer.Option{
				WithPathResolver(ParentLocalPathResolver(dir)),
			},
			false,
			encoded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			gm := goldmark.New(
				goldmark.WithExtensions(Img64),
				goldmark.WithRendererOptions(tt.options...),
			)

			err := gm.Convert([]byte(tt.mdText), &buf)
			gotError := err != nil
			if tt.wantError != gotError {
				t.Errorf("want error %t, got %t", tt.wantError, gotError)
				return
			}
			if err != nil {
				return
			}
			if got := buf.String(); tt.want != got {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}
