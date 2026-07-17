package tsvgoldmark

import (
	"html"
	"strings"

	tsvsheet "github.com/uplang/go-tsvsheet"
)

// sheetHTML parses and computes a .tsvt body and returns its HTML: a computed
// <table> on success, or a visible error <div> when the body is malformed. A
// computed error value in a cell (#DIV/0!, #REF!) is data and renders as its
// cell text within the table.
func (c config) sheetHTML(src []byte) string {
	sheet, err := tsvsheet.Parse(src)
	if err != nil {
		return errorHTML(err)
	}
	return c.tableHTML(sheet.Compute(), src)
}

// errorHTML renders a parse failure as a visible, HTML-escaped error pane so a
// malformed block is never silent and never breaks the page.
func errorHTML(err error) string {
	return `<div class="tsvsheet-error">` + html.EscapeString(err.Error()) + `</div>`
}

// tableHTML renders a computed grid as a <table>, one <tr> per row and one <td>
// per HTML-escaped cell (cell text is untrusted), followed by the optional
// raw-source <details> pane. Builder writes never fail, so their always-nil
// errors are deliberately discarded.
func (c config) tableHTML(grid tsvsheet.Grid, src []byte) string {
	var b strings.Builder
	_, _ = b.WriteString(`<table class="`)
	_, _ = b.WriteString(html.EscapeString(string(c.class)))
	_, _ = b.WriteString(`">`)
	for _, row := range grid {
		_, _ = b.WriteString("<tr>")
		for _, cell := range row {
			_, _ = b.WriteString("<td>")
			_, _ = b.WriteString(html.EscapeString(cell))
			_, _ = b.WriteString("</td>")
		}
		_, _ = b.WriteString("</tr>")
	}
	_, _ = b.WriteString("</table>")
	c.writeSource(&b, src)
	return b.String()
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
