package tsvgoldmark

import (
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// errBoom is the injected write failure the failing writer reports.
const errBoom errs.Const = "boom"

// failWriter is a util.BufWriter whose writes always fail, exercising the
// render write-error path.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error)       { return 0, errBoom }
func (failWriter) Available() int                  { return 0 }
func (failWriter) Buffered() int                   { return 0 }
func (failWriter) Flush() error                    { return errBoom }
func (failWriter) WriteByte(byte) error            { return errBoom }
func (failWriter) WriteRune(rune) (int, error)     { return 0, errBoom }
func (failWriter) WriteString(string) (int, error) { return 0, errBoom }

// firstFenced returns the first fenced code block in a parsed document.
func firstFenced(t *testing.T, src []byte) *ast.FencedCodeBlock {
	t.Helper()
	doc := goldmark.New().Parser().Parse(text.NewReader(src))
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if block, ok := n.(*ast.FencedCodeBlock); ok {
			return block
		}
	}
	t.Fatal("no fenced code block parsed")
	return nil
}

// TestRenderWriteError asserts a failed table write surfaces ErrRender (with the
// underlying cause) rather than being swallowed.
func TestRenderWriteError(t *testing.T) {
	t.Parallel()
	src := []byte("```sheet\n1\t2\n```\n")
	block := firstFenced(t, src)
	status, err := newNodeRenderer(resolve(nil)).render(failWriter{}, src, block, true)
	require.Equal(t, ast.WalkStop, status)
	require.ErrorIs(t, err, ErrRender)
	require.ErrorIs(t, err, errBoom)
}
