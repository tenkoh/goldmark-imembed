package img64

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

const (
	optImg64ParentPath renderer.OptionName = "base64ParentPath"
	optImg64FileReader renderer.OptionName = "base64FileReader"
)

// Img64Config embeds html.Config to refer to some fields like unsafe and xhtml.
type Img64Config struct {
	html.Config
	ParentPath string
	FileReader FileReader
}

// SetOption implements renderer.NodeRenderer.SetOption
func (c *Img64Config) SetOption(name renderer.OptionName, value any) {
	c.Config.SetOption(name, value)

	switch name {
	case optImg64ParentPath:
		c.ParentPath = value.(string)
	case optImg64FileReader:
		c.FileReader = value.(FileReader)
	}
}

type Img64Option interface {
	renderer.Option
	SetImg64Option(*Img64Config)
}

func WithParentPath(path string) interface {
	renderer.Option
	Img64Option
} {
	return &withParentPath{path}
}

type FileReader func(path string) ([]byte, error)

func defaultFileReader(path string) ([]byte, error) {
	// do not encode online image
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return nil, nil
	}
	return os.ReadFile(filepath.Clean(path))
}

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

type withParentPath struct {
	path string
}

func (o *withParentPath) SetConfig(c *renderer.Config) {
	c.Options[optImg64ParentPath] = o.path
}

func (o *withParentPath) SetImg64Option(c *Img64Config) {
	c.ParentPath = o.path
}

type img64Renderer struct {
	Img64Config
}

func NewImg64Renderer(opts ...Img64Option) renderer.NodeRenderer {
	r := &img64Renderer{
		Img64Config: Img64Config{
			Config:     html.NewConfig(),
			FileReader: defaultFileReader,
		},
	}
	for _, o := range opts {
		o.SetImg64Option(&r.Img64Config)
	}
	return r
}

func (r *img64Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

// see https://developer.mozilla.org/ja/docs/Web/Media/Formats/Image_types
var commonWebImages = func() map[string]struct{} {
	types := []string{
		"image/apng",
		"image/avif",
		"image/gif",
		"image/jpeg",
		"image/png",
		"image/svg+xml",
		"image/webp",
	}
	m := map[string]struct{}{}
	for _, t := range types {
		m[t] = struct{}{}
	}
	return m
}()

func (r *img64Renderer) encodeImage(src []byte) ([]byte, error) {
	s := string(src)
	// already encoded
	if strings.HasPrefix(s, "data:") {
		return src, nil
	}
	if !filepath.IsAbs(s) && r.ParentPath != "" {
		s = filepath.Join(r.ParentPath, s)
	}
	b, err := r.FileReader(s)
	if err != nil {
		return nil, fmt.Errorf("fail to read %s: %w", s, err)
	}
	if b == nil {
		return src, nil // do not encode unsupported images
	}
	mtype := mimetype.Detect(b).String()
	if _, exist := commonWebImages[mtype]; !exist {
		return nil, fmt.Errorf("can not embed the filetype %s", mtype)
	}

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("data:%s;base64,", mtype)))
	enc := base64.NewEncoder(base64.StdEncoding, &buf)
	enc.Write(b)
	enc.Close()

	return buf.Bytes(), nil
}

// renderImage adds image embedding function to github.com/yuin/goldmark/renderer/html (MIT).
func (r *img64Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString(`<img src="`)
	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		s, err := r.encodeImage(n.Destination)
		if err != nil || s == nil {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		} else {
			_, _ = w.Write(s)
		}
	}
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n, source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	if r.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Text(source))
		} else if !c.HasChildren() {
			buf.Write(util.EscapeHTML(c.Text(source)))
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}

// img64 implements goldmark.Extender
type img64 struct {
	options []Img64Option
}

// Img64 is an implementation of goldmark.Extender
var Img64 = &img64{}

// NewImg64 initializes Img64: goldmark's extension with its options.
// Using default Img64 with goldmark.WithRendereOptions(opts) give the same result.
func NewImg64(opts ...Img64Option) goldmark.Extender {
	return &img64{
		options: opts,
	}
}

func (e *img64) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewImg64Renderer(e.options...), 500),
	))
}
