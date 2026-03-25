// Package beancount provides parsing and syntax highlighting for beancount files
// using gotreesitter's pure-Go tree-sitter runtime.
package beancount

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"codeberg.org/hum3/gotreesitter"
	"codeberg.org/hum3/gotreesitter/grammars"
)

var (
	blankLine      = []byte("\n\n")
	beancountEntry = grammars.DetectLanguage(".beancount")
)

// Language returns the beancount tree-sitter language with external scanner attached.
func Language() *gotreesitter.Language {
	lang := beancountEntry.Language()
	lang.ExternalScanner = grammars.BeancountExternalScanner{}
	return lang
}

// StripBlankLines collapses consecutive blank lines to a single newline.
// Workaround for gotreesitter beancount grammar bug where 2+ blank lines
// in the input cause the parse tree root to become an error node.
// Parse and Highlight call this automatically; use it when you need byte
// offsets from those functions to align with the source text.
func StripBlankLines(src []byte) []byte {
	for bytes.Contains(src, blankLine) {
		src = bytes.ReplaceAll(src, blankLine, []byte("\n"))
	}
	return src
}

// Parse parses beancount source and returns the syntax tree.
func Parse(src []byte) (*gotreesitter.Tree, error) {
	parser := gotreesitter.NewParser(Language())
	tree, err := parser.Parse(StripBlankLines(src))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return tree, nil
}

// ParseFile reads a file and parses it, returning a BoundTree for convenient access.
func ParseFile(path string) (*gotreesitter.BoundTree, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	tree, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return gotreesitter.Bind(tree), nil
}

// HasErrors reports whether the tree contains any parse errors.
func HasErrors(tree *gotreesitter.Tree) bool {
	return tree.RootNode() != nil && tree.RootNode().HasError()
}

// SExpression returns an S-expression string representation of the parse tree.
func SExpression(tree *gotreesitter.Tree) string {
	if tree.RootNode() == nil {
		return "()"
	}
	var buf strings.Builder
	writeNode(&buf, tree.RootNode(), tree.Language(), 0)
	return buf.String()
}

func writeNode(buf *strings.Builder, node *gotreesitter.Node, lang *gotreesitter.Language, depth int) {
	nodeType := node.Type(lang)
	if !node.IsNamed() {
		return
	}
	buf.WriteString(strings.Repeat("  ", depth))
	buf.WriteString("(")
	buf.WriteString(nodeType)
	childCount := node.ChildCount()
	hasNamedChildren := false
	for i := range childCount {
		child := node.Child(i)
		if child.IsNamed() {
			hasNamedChildren = true
			break
		}
	}
	if hasNamedChildren {
		buf.WriteString("\n")
		for i := range childCount {
			child := node.Child(i)
			if child.IsNamed() {
				writeNode(buf, child, lang, depth+1)
			}
		}
		buf.WriteString(strings.Repeat("  ", depth))
	}
	buf.WriteString(")\n")
}

// CollectNodeTypes walks the tree and returns a set of all named node types present.
func CollectNodeTypes(tree *gotreesitter.Tree) map[string]bool {
	types := make(map[string]bool)
	if tree.RootNode() == nil {
		return types
	}
	collectTypes(tree.RootNode(), tree.Language(), types)
	return types
}

func collectTypes(node *gotreesitter.Node, lang *gotreesitter.Language, types map[string]bool) {
	if node.IsNamed() {
		types[node.Type(lang)] = true
	}
	for i := 0; i < node.ChildCount(); i++ {
		collectTypes(node.Child(i), lang, types)
	}
}
