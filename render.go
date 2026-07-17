package tsvgoldmark

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"

	"github.com/tsvsheet/tsvsheet.goldmark/internal/constants"
)

// sheetLang is the info-string language a fenced code block must declare to be
// rendered as a computed table.
var sheetLang = []byte("sheet")

// nodeRenderer renders fenced code blocks: ```sheet blocks become computed
// tables, and every other language delegates to fallback (goldmark's default
// FencedCodeBlock renderer).
type nodeRenderer struct {
	fallback renderer.NodeRendererFunc
	cfg      config
}

// newNodeRenderer builds the fenced-code-block renderer, capturing goldmark's
// default renderer as the fallback for non-sheet languages.
func newNodeRenderer(cfg config) nodeRenderer {
	return nodeRenderer{cfg: cfg, fallback: defaultFencedFunc()}
}

// RegisterFuncs claims FencedCodeBlock rendering for this renderer.
func (r nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.render)
}

// render dispatches one fenced code block: a `sheet` block is computed and
// tabulated on entry (its text children skipped); any other language is handed
// to goldmark's default renderer unchanged.
func (r nodeRenderer) render(
	w util.BufWriter, source []byte, node ast.Node, isEntering bool,
) (ast.WalkStatus, error) {
	// Only KindFencedCodeBlock is registered to this renderer, so the assertion
	// is total.
	block := node.(*ast.FencedCodeBlock)
	if !bytes.Equal(block.Language(source), sheetLang) {
		return r.fallback(w, source, node, isEntering)
	}
	if !isEntering {
		return ast.WalkContinue, nil
	}
	if _, err := w.WriteString(r.cfg.sheetHTML(blockText(block, source))); err != nil {
		return ast.WalkStop, constants.ErrRender.With(err)
	}
	return ast.WalkSkipChildren, nil
}

// blockText concatenates the raw lines of a fenced code block into its .tsvt
// body, preserving the TAB and newline bytes the engine parses.
func blockText(block *ast.FencedCodeBlock, source []byte) []byte {
	lines := block.Lines()
	var buf bytes.Buffer
	for i := range lines.Len() {
		seg := lines.At(i)
		_, _ = buf.Write(seg.Value(source))
	}
	return buf.Bytes()
}

// defaultFencedFunc returns goldmark's default FencedCodeBlock render function,
// captured from a stock HTML renderer so non-sheet blocks render identically to
// an unextended goldmark.
func defaultFencedFunc() renderer.NodeRendererFunc {
	reg := capture{}
	html.NewRenderer().RegisterFuncs(reg)
	return reg[ast.KindFencedCodeBlock]
}

// capture is a NodeRendererFuncRegisterer that records the render function of
// every node kind a renderer registers, keyed by kind. It is a map (a reference
// type), so a value receiver mutates the shared backing store.
type capture map[ast.NodeKind]renderer.NodeRendererFunc

// Register records fn under kind.
func (c capture) Register(kind ast.NodeKind, fn renderer.NodeRendererFunc) {
	c[kind] = fn
}
