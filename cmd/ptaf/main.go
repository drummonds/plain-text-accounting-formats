package main

import (
	"fmt"
	"os"
	"sort"

	"codeberg.org/hum3/gotreesitter"
	beancount "codeberg.org/hum3/plain-text-accounting-formats"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: ptaf <parse|highlight|check|convert> <file>\n")
		os.Exit(1)
	}
	cmd, path := os.Args[1], os.Args[2]

	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "convert":
		out, err := beancount.Convert(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, _ = os.Stdout.Write(out)

	case "parse":
		tree, err := beancount.Parse(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(beancount.SExpression(tree))

	case "highlight":
		ranges, err := beancount.Highlight(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printANSI(beancount.StripBlankLines(src), ranges)

	case "check":
		tree, err := beancount.Parse(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if beancount.HasErrors(tree) {
			fmt.Fprintf(os.Stderr, "%s: parse errors found\n", path)
			printErrors(tree.RootNode(), beancount.Language(), src, 0)
			os.Exit(1)
		}
		fmt.Printf("%s: OK\n", path)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: ptaf <parse|highlight|check|convert> <file>\n", cmd)
		os.Exit(1)
	}
}

// ANSI color codes for highlight captures.
var ansiColors = map[string]string{
	"keyword":          "\033[1;35m", // bold magenta
	"string":           "\033[32m",   // green
	"number":           "\033[33m",   // yellow
	"constant":         "\033[36m",   // cyan
	"constant.builtin": "\033[1;36m", // bold cyan
	"type":             "\033[1;34m", // bold blue
	"tag":              "\033[34m",   // blue
	"attribute":        "\033[34m",   // blue
	"comment":          "\033[2;37m", // dim white
	"operator":         "\033[31m",   // red
}

const ansiReset = "\033[0m"

func printANSI(src []byte, ranges []gotreesitter.HighlightRange) {
	// Sort ranges by start byte
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].StartByte < ranges[j].StartByte
	})

	pos := uint32(0)
	for _, r := range ranges {
		// Print unhighlighted gap
		if r.StartByte > pos {
			_, _ = os.Stdout.Write(src[pos:r.StartByte])
		}
		// Print highlighted range
		color, ok := ansiColors[r.Capture]
		if !ok {
			color = ""
		}
		if color != "" {
			_, _ = os.Stdout.WriteString(color)
		}
		_, _ = os.Stdout.Write(src[r.StartByte:r.EndByte])
		if color != "" {
			_, _ = os.Stdout.WriteString(ansiReset)
		}
		pos = r.EndByte
	}
	// Print remaining
	if pos < uint32(len(src)) {
		_, _ = os.Stdout.Write(src[pos:])
	}
}

func printErrors(node *gotreesitter.Node, lang *gotreesitter.Language, src []byte, depth int) {
	if node == nil {
		return
	}
	if node.HasError() && node.IsNamed() {
		nodeType := node.Type(lang)
		if nodeType == "ERROR" || node.IsMissing() {
			start := node.StartPoint()
			fmt.Fprintf(os.Stderr, "  line %d:%d: %s\n", start.Row+1, start.Column, nodeType)
		}
	}
	for i := range node.ChildCount() {
		printErrors(node.Child(i), lang, src, depth+1)
	}
}
