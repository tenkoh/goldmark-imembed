package img64

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/renderer"
)

type FileReader func(path string) ([]byte, error)

func DefaultFileReader() FileReader {
	return defaultFileReader
}

func defaultFileReader(path string) ([]byte, error) {
	// do not encode online image
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return nil, nil
	}
	return os.ReadFile(filepath.Clean(path))
}

// WithFileReader changes reading image attribute from default.
// For example, it allows reading online images. (it is disabled in default reader)
func WithFileReader(r FileReader) interface {
	renderer.Option
	Img64Option
} {
	return &withFileReader{r}
}

type withFileReader struct {
	reader FileReader
}

func (o *withFileReader) SetConfig(c *renderer.Config) {
	c.Options[optImg64FileReader] = o.reader
}

func (o *withFileReader) SetImg64Option(c *Img64Config) {
	c.FileReader = o.reader
}

// AllowRemoteFileReader enables embedding remote images which is prohibited with default reader.
// Use this only if all images are confirmed to be safe.
// This is one example of FileReader implementations.
func AllowRemoteFileReader(client *http.Client) FileReader {
	return func(path string) ([]byte, error) {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			resp, err := client.Get(path)
			if err != nil {
				return nil, fmt.Errorf("fail to get image from %s: %w", path, err)
			}
			defer resp.Body.Close()
			return io.ReadAll(resp.Body)
		}
		return os.ReadFile(path)
	}
}
