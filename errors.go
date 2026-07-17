package tsvgoldmark

import "github.com/tsvsheet/tsvsheet.goldmark/internal/constants"

// ErrRender is returned by the goldmark render pass when the computed table HTML
// cannot be written to the output stream; it is matchable with errors.Is and
// wraps the underlying write failure as its cause. A malformed .tsvt block never
// produces this error — it renders an inline error <div> and the conversion
// succeeds.
const ErrRender = constants.ErrRender
