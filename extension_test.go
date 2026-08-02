package tsvgoldmark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"

	tsvgoldmark "github.com/tsvsheet/tsvsheet.goldmark"
)

// fence wraps a .tsvt body in a fenced code block with the given info string.
func fence(info, body string) string {
	return "```" + info + "\n" + body + "\n```\n"
}

// convert renders markdown through a goldmark built with the given extender.
func convert(t *testing.T, ext goldmark.Extender, src string) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, goldmark.New(goldmark.WithExtensions(ext)).Convert([]byte(src), &buf))
	return buf.String()
}

func TestConvert(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ext     goldmark.Extender
		src     string
		want    []string
		notWant []string
	}{
		{
			name: "sheet block computes and tabulates",
			ext:  tsvgoldmark.New(),
			src:  fence("sheet", "1\t2\t=A1+B1"),
			want: []string{`<table class="tsvsheet">`, "<tr>", "<td>1</td>", "<td>2</td>", "<td>3</td>", "</table>"},
		},
		{
			name: "a declared header row renders as th",
			ext:  tsvgoldmark.New(),
			src:  fence("sheet", "#.header\trows(count(1))\nname\tqty\nwidget\t3"),
			want: []string{"<th>name</th>", "<th>qty</th>", "<td>widget</td>"},
		},
		{
			name:    "a hidden column is left out of the table",
			ext:     tsvgoldmark.New(),
			src:     fence("sheet", "#.hide\tcols(range(B:B))\nname\tscratch\nwidget\tx"),
			want:    []string{"<td>name</td>", "<td>widget</td>"},
			notWant: []string{"scratch", "<td>x</td>"},
		},
		{
			name:    "a hidden row is left out of the table",
			ext:     tsvgoldmark.New(),
			src:     fence("sheet", "#.hide\trows(range(2:2))\nkeep\ndrop\nkeep too"),
			want:    []string{"<td>keep</td>", "<td>keep too</td>"},
			notWant: []string{"<td>drop</td>"},
		},
		{
			name:    "non-sheet fence renders normally",
			ext:     tsvgoldmark.New(),
			src:     fence("go", "fmt.Println(1)"),
			want:    []string{`<code class="language-go">`, "fmt.Println(1)"},
			notWant: []string{"<table"},
		},
		{
			name:    "plain fence with no language renders normally",
			ext:     tsvgoldmark.New(),
			src:     fence("", "just text"),
			want:    []string{"<pre><code>", "just text"},
			notWant: []string{"<table"},
		},
		{
			name:    "malformed sheet yields a visible error div",
			ext:     tsvgoldmark.New(),
			src:     fence("sheet", "=1+"),
			want:    []string{`<div class="tsvsheet-error">`, "syntax error"},
			notWant: []string{"<table"},
		},
		{
			name: "computed cell error value renders as text",
			ext:  tsvgoldmark.New(),
			src:  fence("sheet", "10\t0\t=A1/B1"),
			want: []string{"<td>#DIV/0!</td>"},
		},
		{
			name:    "html in a cell is escaped",
			ext:     tsvgoldmark.New(),
			src:     fence("sheet", "<b>x</b>"),
			want:    []string{"<td>&lt;b&gt;x&lt;/b&gt;</td>"},
			notWant: []string{"<td><b>x</b></td>"},
		},
		{
			name: "WithClass changes the table class",
			ext:  tsvgoldmark.New(tsvgoldmark.WithClass("grid")),
			src:  fence("sheet", "1\t2"),
			want: []string{`<table class="grid">`},
		},
		{
			name: "WithSource appends the raw source pane",
			ext:  tsvgoldmark.New(tsvgoldmark.WithSource(true)),
			src:  fence("sheet", "1\t2"),
			want: []string{`<details class="tsvsheet-source">`, "<summary>source</summary>", "1\t2"},
		},
		{
			name:    "default omits the source pane",
			ext:     tsvgoldmark.New(),
			src:     fence("sheet", "1\t2"),
			notWant: []string{"<details"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := convert(t, tc.ext, tc.src)
			for _, w := range tc.want {
				require.Contains(t, got, w)
			}
			for _, nw := range tc.notWant {
				require.NotContains(t, got, nw)
			}
		})
	}
}

// TestConvertDoesNotPanicOnMalformed guards the "errors are visible, never a
// panic" contract independently of substring assertions.
func TestConvertDoesNotPanicOnMalformed(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() {
		_ = convert(t, tsvgoldmark.New(), fence("sheet", "=BOGUS("))
	})
}

// TestWithClassEmpty confirms an empty class name renders valid, empty markup.
func TestWithClassEmpty(t *testing.T) {
	t.Parallel()
	got := convert(t, tsvgoldmark.New(tsvgoldmark.WithClass("")), fence("sheet", "1"))
	require.Contains(t, got, `<table class="">`)
	require.True(t, strings.Contains(got, "<td>1</td>"))
}

// TestErrRenderIsNotWhatAMalformedBlockProduces pins the difference between the
// two ways rendering can go wrong. A broken .tsvt block is the author's
// problem and belongs on the page as a visible pane; a write failure is the
// caller's problem and belongs in the error return. Conflating them would
// either fail a whole document over one bad fence, or swallow a genuine I/O
// failure as if it were content.
func TestErrRenderIsNotWhatAMalformedBlockProduces(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := goldmark.New(goldmark.WithExtensions(tsvgoldmark.New())).
		Convert([]byte(fence("sheet", "=1+")), &buf)

	require.NoError(t, err, "a malformed block is content, not a conversion failure")
	assert.NotErrorIs(t, err, tsvgoldmark.ErrRender)
	assert.Contains(t, buf.String(), "<div", "it renders as a pane on the page")
}

// TestErrorHTMLMakesAMalformedBlockVisibleWithoutBreakingThePage pins both
// halves: the failure is shown rather than silently dropped, and it is escaped
// rather than injected — a parse error carrying markup from the source would
// turn a typo into an XSS vector.
func TestErrorHTMLMakesAMalformedBlockVisibleWithoutBreakingThePage(t *testing.T) {
	t.Parallel()
	out := convert(t, tsvgoldmark.New(), fence("sheet", "=<script>alert(1)</script>"))

	assert.Contains(t, out, `<div class="tsvsheet-error">`, "the failure is never silent")
	assert.NotContains(t, out, "<script", "and never breaks the page it lands on")
	assert.Contains(t, out, "&lt;", "the offending text is escaped into the message")
}

// TestTableHTMLEscapesUntrustedCellText pins the escaping. Cell text comes from
// a document, and a document is untrusted input; rendering it raw would make
// every published sheet a way to inject markup into the page around it.
func TestTableHTMLEscapesUntrustedCellText(t *testing.T) {
	t.Parallel()
	out := convert(t, tsvgoldmark.New(), fence("sheet", "<img src=x onerror=alert(1)>\tplain"))

	assert.NotContains(t, out, "<img", "the cell's markup is text, not markup")
	assert.Contains(t, out, "&lt;img")
	assert.Contains(t, out, "plain")
}

// TestExtensionIsSafeToShareAcrossGoldmarkInstances pins the immutability the
// doc claims. An extender that accumulated per-conversion state would work
// perfectly until someone reused it, and then leak one document's content into
// another's — the kind of bug that only appears under concurrency or in a
// long-lived server.
func TestExtensionIsSafeToShareAcrossGoldmarkInstances(t *testing.T) {
	t.Parallel()
	shared := tsvgoldmark.New()

	first := convert(t, shared, fence("sheet", "1\t2"))
	second := convert(t, shared, fence("sheet", "3\t4"))
	again := convert(t, shared, fence("sheet", "1\t2"))

	assert.Equal(t, first, again, "the same input renders the same however often the extender is reused")
	assert.NotEqual(t, first, second)
	assert.NotContains(t, second, ">1<", "and no document leaks into the next")
}
