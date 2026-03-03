package beancount

import (
	"testing"
)

const golucaSample = `2024-01-01 *
  Equity:Opening -> Assets:Bank "Opening balance" 1000.00 GBP
2024-01-15 *
  Assets:Bank -> Expenses:Groceries "Weekly groceries" 45.50 GBP
`

func TestGolucaParse(t *testing.T) {
	tree, err := GolucaParse([]byte(golucaSample))
	if err != nil {
		t.Fatal(err)
	}
	root := tree.RootNode()
	if root == nil {
		t.Fatal("root node is nil")
	}
	typ := root.Type(GolucaLanguage())
	if typ != "source_file" {
		t.Fatalf("root type = %q, want %q", typ, "source_file")
	}
	if root.HasError() {
		t.Error("unexpected parse errors")
	}
}

func TestGolucaHighlight(t *testing.T) {
	ranges, err := GolucaHighlight([]byte(golucaSample))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) == 0 {
		t.Fatal("expected highlight ranges")
	}
	captures := make(map[string]bool)
	for _, r := range ranges {
		captures[r.Capture] = true
	}
	for _, want := range []string{"constant", "keyword", "type", "string", "number"} {
		if !captures[want] {
			t.Errorf("missing capture type %q", want)
		}
	}
}

func TestConvertThenHighlight(t *testing.T) {
	src := []byte(`2024-01-01 * "Shop" "Groceries"
  Expenses:Food   25.00 GBP
  Assets:Cash
`)
	goluca, err := Convert(src)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if len(goluca) == 0 {
		t.Fatal("convert produced empty output")
	}

	tree, err := GolucaParse(goluca)
	if err != nil {
		t.Fatalf("goluca parse: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("goluca parse errors after conversion")
	}

	ranges, err := GolucaHighlight(goluca)
	if err != nil {
		t.Fatalf("goluca highlight: %v", err)
	}
	if len(ranges) == 0 {
		t.Fatal("expected highlight ranges from converted goluca")
	}
}
