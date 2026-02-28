package beancount

import (
	"os"
	"strings"
	"testing"
)

func TestParseSimple(t *testing.T) {
	src, err := os.ReadFile("testdata/simple.beancount")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	root := tree.RootNode()
	if root == nil {
		t.Fatal("root node is nil")
	}
	typ := root.Type(Language())
	if typ != "file" {
		t.Fatalf("root type = %q, want %q", typ, "file")
	}
	if root.ChildCount() == 0 {
		t.Fatal("expected children in root node")
	}
}

func TestParseTransactions(t *testing.T) {
	src, err := os.ReadFile("testdata/transactions.beancount")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	types := CollectNodeTypes(tree)
	for _, want := range []string{"transaction", "posting", "amount"} {
		if !types[want] {
			t.Errorf("missing node type %q", want)
		}
	}
}

func TestParseFull(t *testing.T) {
	src, err := os.ReadFile("testdata/full.beancount")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	types := CollectNodeTypes(tree)

	// All 12 entry types
	entryTypes := []string{
		"transaction", "balance", "open", "close", "pad",
		"document", "note", "event", "price", "commodity",
		"query", "custom",
	}
	for _, want := range entryTypes {
		if !types[want] {
			t.Errorf("missing entry type %q", want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	src := []byte("2024-01-01 open\n")
	tree, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if !HasErrors(tree) {
		t.Error("expected parse errors for malformed input")
	}
}

func TestHighlight(t *testing.T) {
	src, err := os.ReadFile("testdata/simple.beancount")
	if err != nil {
		t.Fatal(err)
	}
	ranges, err := Highlight(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) == 0 {
		t.Fatal("expected highlight ranges")
	}

	// Check that we got a variety of capture types
	captures := make(map[string]bool)
	for _, r := range ranges {
		captures[r.Capture] = true
	}
	for _, want := range []string{"keyword", "string", "number", "type", "constant", "comment"} {
		if !captures[want] {
			t.Errorf("missing capture type %q", want)
		}
	}
}

func TestSExpression(t *testing.T) {
	src, err := os.ReadFile("testdata/simple.beancount")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	sexpr := SExpression(tree)
	if !strings.HasPrefix(sexpr, "(file\n") {
		t.Errorf("S-expression should start with (file, got: %s", sexpr[:min(60, len(sexpr))])
	}
	if !strings.Contains(sexpr, "(transaction") {
		t.Error("S-expression should contain (transaction")
	}
	if !strings.Contains(sexpr, "(open") {
		t.Error("S-expression should contain (open")
	}
}

func TestParseFile(t *testing.T) {
	bt, err := ParseFile("testdata/simple.beancount")
	if err != nil {
		t.Fatal(err)
	}
	defer bt.Release()
	root := bt.RootNode()
	if root == nil {
		t.Fatal("root is nil")
	}
	if bt.NodeType(root) != "file" {
		t.Fatalf("root type = %q, want %q", bt.NodeType(root), "file")
	}
}

func TestParseRealFile(t *testing.T) {
	path := os.Getenv("BEANCOUNT_TEST_FILE")
	if path == "" {
		t.Skip("set BEANCOUNT_TEST_FILE to test with a real beancount file")
	}
	bt, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer bt.Release()
	if bt.RootNode() == nil {
		t.Fatal("root is nil")
	}
}
