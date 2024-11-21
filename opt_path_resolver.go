package img64

import (
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/renderer"
)

type PathResolver func(string) string

// WithPathResolver adds custom behavior to read image source path.
// For example, relative paths could be converted into absolute paths by adding the parent directory path.
func WithPathResolver(r PathResolver) interface {
	renderer.Option
	Img64Option
} {
	return &withPathResolver{r}
}

type withPathResolver struct {
	r PathResolver
}

func (o *withPathResolver) SetConfig(c *renderer.Config) {
	c.Options[optImg64PathResolver] = o.r
}

func (o *withPathResolver) SetImg64Option(c *Img64Config) {
	c.PathResolver = o.r
}

func DefaultPathResolver() PathResolver {
	return defaultPathResolver
}

func defaultPathResolver(path string) string {
	return path
}

// ParentLocalPathResolver adds parent directory path (ex. /var + target.md = /var/target.md).
// This is one example of PathResolver implementations.
func ParentLocalPathResolver(parentPath string) PathResolver {
	return func(path string) string {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return path
		}
		return filepath.Join(parentPath, path)
	}
}
