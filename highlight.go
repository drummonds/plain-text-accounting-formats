package beancount

import (
	"fmt"

	"github.com/drummonds/gotreesitter"
)

// Highlight parses beancount source and returns highlight ranges.
func Highlight(src []byte) ([]gotreesitter.HighlightRange, error) {
	hl, err := gotreesitter.NewHighlighter(Language(), beancountEntry.HighlightQuery)
	if err != nil {
		return nil, fmt.Errorf("highlight query: %w", err)
	}
	return hl.Highlight(StripBlankLines(src)), nil
}
