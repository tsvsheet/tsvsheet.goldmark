package tsvgoldmark

// ClassName is the CSS class placed on the rendered <table> element. It names a
// stylesheet hook the host controls, not untrusted cell data.
type ClassName string

// ShowSource selects whether the raw .tsvt source of a `sheet` block is emitted
// alongside the computed table in a collapsible <details> pane.
type ShowSource bool

// defaultClass is the <table> class applied when WithClass is not supplied.
const defaultClass ClassName = "tsvsheet"

// config is the resolved, immutable rendering configuration a set of Options
// produces. It is copied by value into the node renderer.
type config struct {
	class         ClassName
	sourceEnabled bool
}

// Option configures the extension's rendering. Options are applied in order by
// New; a later Option overrides an earlier one for the same setting.
type Option func(config) config

// WithClass sets the CSS class on the rendered <table>. The default is
// "tsvsheet". An empty class renders `class=""`, which is valid HTML.
func WithClass(name ClassName) Option {
	return func(c config) config {
		c.class = name
		return c
	}
}

// WithSource controls whether each rendered `sheet` block also emits its raw
// .tsvt source in a collapsible <details> pane after the table. The default is
// false — table only.
func WithSource(shouldShow ShowSource) Option {
	return func(c config) config {
		c.sourceEnabled = bool(shouldShow)
		return c
	}
}

// resolve folds the options over the defaults into an immutable config.
func resolve(opts []Option) config {
	c := config{class: defaultClass}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}
