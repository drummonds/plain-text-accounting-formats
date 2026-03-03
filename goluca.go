package beancount

import (
	"fmt"

	"github.com/drummonds/gotreesitter"
	"github.com/drummonds/gotreesitter/grammars"
)

var golucaEntry = grammars.DetectLanguage(".goluca")

// GolucaLanguage returns the goluca tree-sitter language.
func GolucaLanguage() *gotreesitter.Language {
	return golucaEntry.Language()
}

// GolucaParse parses goluca source and returns the syntax tree.
func GolucaParse(src []byte) (*gotreesitter.Tree, error) {
	parser := gotreesitter.NewParser(GolucaLanguage())
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("goluca parse: %w", err)
	}
	return tree, nil
}

// GolucaHighlight parses goluca source and returns highlight ranges.
func GolucaHighlight(src []byte) ([]gotreesitter.HighlightRange, error) {
	hl, err := gotreesitter.NewHighlighter(GolucaLanguage(), golucaEntry.HighlightQuery)
	if err != nil {
		return nil, fmt.Errorf("goluca highlight query: %w", err)
	}
	return hl.Highlight(src), nil
}
