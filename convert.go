package beancount

import (
	"fmt"
	"strings"

	"github.com/drummonds/gotreesitter"
)

// Convert transforms beancount source into go-luca movement format.
// Only 2-posting transactions are converted; N-posting transactions are
// skipped with a comment, and non-transaction directives are dropped.
func Convert(src []byte) ([]byte, error) {
	tree, err := Parse(src)
	if err != nil {
		return nil, err
	}
	bt := gotreesitter.Bind(tree)
	defer bt.Release()

	root := bt.RootNode()
	if root == nil {
		return nil, fmt.Errorf("empty parse tree")
	}

	lang := Language()
	var out strings.Builder

	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if !child.IsNamed() {
			continue
		}
		nodeType := child.Type(lang)
		switch nodeType {
		case "transaction":
			convertTransaction(&out, bt, lang, child)
		case "comment":
			// drop silently
		default:
			fmt.Fprintf(&out, "; DROPPED: %s\n", nodeType)
		}
	}

	return []byte(out.String()), nil
}

type postingInfo struct {
	account   string
	amount    string // empty if auto-balanced
	commodity string
	negative  bool
}

func convertTransaction(out *strings.Builder, bt *gotreesitter.BoundTree, lang *gotreesitter.Language, node *gotreesitter.Node) {
	var date, flag, payee, narration string
	var postings []*gotreesitter.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if !child.IsNamed() {
			continue
		}
		switch child.Type(lang) {
		case "date":
			date = bt.NodeText(child)
		case "txn":
			flag = bt.NodeText(child)
		case "payee":
			payee = unquote(bt.NodeText(child))
		case "narration":
			narration = unquote(bt.NodeText(child))
		case "posting":
			postings = append(postings, child)
		}
	}

	desc := narration
	if payee != "" {
		desc = payee + " \u2014 " + narration
	}

	if len(postings) != 2 {
		fmt.Fprintf(out, "; SKIPPED (%d postings): %s\n", len(postings), desc)
		return
	}

	p1 := extractPosting(bt, lang, postings[0])
	p2 := extractPosting(bt, lang, postings[1])
	from, to := determineFromTo(p1, p2)

	fmt.Fprintf(out, "%s %s\n", date, flag)
	fmt.Fprintf(out, "  %s -> %s %q %s %s\n", from.account, to.account, desc, to.amount, to.commodity)
}

func extractPosting(bt *gotreesitter.BoundTree, lang *gotreesitter.Language, node *gotreesitter.Node) postingInfo {
	var p postingInfo
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if !child.IsNamed() {
			continue
		}
		switch child.Type(lang) {
		case "account":
			p.account = bt.NodeText(child)
		case "incomplete_amount":
			extractAmount(bt, lang, child, &p)
		}
	}
	return p
}

func extractAmount(bt *gotreesitter.BoundTree, lang *gotreesitter.Language, node *gotreesitter.Node, p *postingInfo) {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if !child.IsNamed() {
			continue
		}
		switch child.Type(lang) {
		case "number":
			p.amount = bt.NodeText(child)
		case "currency":
			p.commodity = bt.NodeText(child)
		case "unary_number_expr":
			for j := 0; j < child.ChildCount(); j++ {
				gc := child.Child(j)
				if gc.IsNamed() {
					switch gc.Type(lang) {
					case "minus":
						p.negative = true
					case "number":
						p.amount = bt.NodeText(gc)
					}
				}
			}
		}
	}
}

// determineFromTo identifies which posting is the source (from) and destination (to).
// Posting with positive explicit amount = to; negative or missing amount = from.
func determineFromTo(a, b postingInfo) (from, to postingInfo) {
	switch {
	case a.amount != "" && !a.negative && (b.amount == "" || b.negative):
		return b, a
	case b.amount != "" && !b.negative && (a.amount == "" || a.negative):
		return a, b
	default:
		// Both positive or ambiguous: first = to, second = from
		return b, a
	}
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
