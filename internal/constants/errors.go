// Package constants declares the extension's sentinel error values. The error
// mechanism (the matchable string type) lives in the shared gomatic/go-error
// library; these values are this package's own.
package constants

// Imported bare (the package is named error); this file declares only sentinels
// and uses no builtin error type, so each declaration reads errs.Const.
import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	// ErrRender is returned when the rendered table HTML cannot be written to
	// the output stream. It wraps the underlying write failure as its cause.
	ErrRender errs.Const = "failed to render sheet"
)
