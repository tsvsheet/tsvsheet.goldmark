package tsvgoldmark_test

import (
	"bytes"
	"strings"
	"testing"

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
