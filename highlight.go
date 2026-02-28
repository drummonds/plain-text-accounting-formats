package beancount

import (
	"fmt"

	"github.com/odvcencio/gotreesitter"
)

// HighlightQuery is the tree-sitter highlight query for beancount files.
const HighlightQuery = `
; Keywords — entry type keywords
[
  "open"
  "close"
  "balance"
  "pad"
  "event"
  "query"
  "note"
  "document"
  "custom"
  "commodity"
  "price"
  "txn"
  "pushtag"
  "poptag"
  "pushmeta"
  "popmeta"
  "option"
  "include"
  "plugin"
] @keyword

; Strings
(string) @string

; Numbers
(number) @number

; Dates
(date) @constant

; Accounts
(account) @type

; Currencies
(currency) @constant

; Tags and links
(tag) @tag
(link) @attribute

; Comments
(comment) @comment

; Flags
(flag) @keyword

; Booleans
(bool) @constant.builtin

; Operators
[
  (plus)
  (minus)
  (asterisk)
  (slash)
  (at)
  (atat)
] @operator
`

// Highlight parses beancount source and returns highlight ranges.
func Highlight(src []byte) ([]gotreesitter.HighlightRange, error) {
	hl, err := gotreesitter.NewHighlighter(Language(), HighlightQuery)
	if err != nil {
		return nil, fmt.Errorf("highlight query: %w", err)
	}
	return hl.Highlight(src), nil
}
