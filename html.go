package tsvgoldmark

import (
	"html"
	"strings"

	tsvsheet "github.com/tsvsheet/go-tsvsheet"
)

// sheetHTML parses and computes a .tsvt body and returns its HTML: a computed
// <table> on success, or a visible error <div> when the body is malformed. A
// computed error value in a cell (#DIV/0!, #REF!) is data and renders as its
// cell text within the table.
func (c config) sheetHTML(src []byte) string {
	doc, err := tsvsheet.ParseDocument(src)
	if err != nil {
		return errorHTML(err)
	}
	view, _ := doc.View()
	return c.tableHTML(doc.Sheet().Compute(), view, src)
}

// errorHTML renders a parse failure as a visible, HTML-escaped error pane so a
// malformed block is never silent and never breaks the page.
func errorHTML(err error) string {
	return `<div class="tsvsheet-error">` + html.EscapeString(err.Error()) + `</div>`
}

// tableHTML renders a computed grid as a <table>, honouring the view the sheet
// declares for itself: a row or column its `#.hide` directives hide is left
// out, and a `#.header` row is written as <th> — a rendered fence is a viewport,
// so the declarations apply. Cell text is untrusted and HTML-escaped. The
// optional raw-source <details> pane follows, carrying the source verbatim so
// what is hidden is still one click away. Builder writes never fail, so their
// always-nil errors are deliberately discarded.
func (c config) tableHTML(grid tsvsheet.Grid, view tsvsheet.View, src []byte) string {
	var b strings.Builder
	_, _ = b.WriteString(`<table class="`)
	_, _ = b.WriteString(html.EscapeString(string(c.class)))
	_, _ = b.WriteString(`">`)
	for r, row := range grid {
		if view.HiddenRows[r+1] {
			continue
		}
		writeRow(&b, row, view, cellTag(isHeaderRow(view.HeaderRows[r+1])))
	}
	_, _ = b.WriteString("</table>")
	c.writeSource(&b, src)
	return b.String()
}

// cellElement is the HTML element one row's cells are written with: th or td.
type cellElement string

// writeRow renders one visible row, dropping the hidden columns.
func writeRow(b *strings.Builder, row []string, view tsvsheet.View, tag cellElement) {
	_, _ = b.WriteString("<tr>")
	for i, cell := range row {
		if view.HiddenCols[i+1] {
			continue
		}
		_, _ = b.WriteString("<" + string(tag) + ">")
		_, _ = b.WriteString(html.EscapeString(cell))
		_, _ = b.WriteString("</" + string(tag) + ">")
	}
	_, _ = b.WriteString("</tr>")
}

// isHeaderRow marks a row the sheet declares as a header, which decides the
// element its cells are written with.
type isHeaderRow bool

// cellTag picks the cell element for a row: <th> for a row the sheet declares
// as a header, <td> otherwise.
func cellTag(isHeader isHeaderRow) cellElement {
	if isHeader {
		return "th"
	}
	return "td"
}

// writeSource appends the raw .tsvt source in a collapsible <details> pane when
// the ShowSource option is set; otherwise it appends nothing.
func (c config) writeSource(b *strings.Builder, src []byte) {
	if !c.sourceEnabled {
		return
	}
	_, _ = b.WriteString(`<details class="tsvsheet-source"><summary>source</summary><pre>`)
	_, _ = b.WriteString(html.EscapeString(string(src)))
	_, _ = b.WriteString("</pre></details>")
}
