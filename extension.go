// Package tsvgoldmark is a goldmark extension that renders a fenced ```sheet
// code block — a .tsvt spreadsheet — to a computed static HTML <table>,
// server-side, through the go-tsvsheet engine. It is pure Go: no JavaScript and
// no WebAssembly.
//
// A ```sheet block's body is parsed and computed by the engine, then tabulated:
// one <tr> per grid row, one <td> per cell, with every cell value HTML-escaped
// because cells are untrusted. A computed error value (#DIV/0!, #REF!) is data
// and renders as its cell text. A malformed .tsvt block renders a visible
// <div class="tsvsheet-error"> rather than a broken page or a panic. Every other
// fenced language falls through to goldmark's default code-block rendering.
package tsvgoldmark

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// fencedPriority registers the sheet renderer ahead of goldmark's default HTML
// renderer (priority 1000). goldmark applies node renderers lowest-priority
// last and last-write-wins, so a value below 1000 overrides FencedCodeBlock
// rendering while every other node keeps its default.
const fencedPriority = 100

// extension is the goldmark.Extender this package provides. It is immutable and
// safe to share across goldmark instances.
type extension struct {
	cfg config
}

// New builds a goldmark.Extender that renders ```sheet blocks as computed HTML
// tables, applying the given options over the defaults. Called with no options
// it uses the defaults: table class "tsvsheet" and no source pane.
func New(opts ...Option) goldmark.Extender {
	return extension{cfg: resolve(opts)}
}

// Extend registers the sheet renderer on m, overriding fenced-code-block
// rendering for the `sheet` language and delegating every other language to
// goldmark's default renderer.
func (e extension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newNodeRenderer(e.cfg), fencedPriority),
	))
}
