package beancount

import (
	"strings"
	"testing"
)

func TestConvertAutoBalanced(t *testing.T) {
	src := []byte(`2024-01-15 * "Tesco" "Weekly groceries"
  Expenses:Groceries   45.50 GBP
  Assets:Bank:Current
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	want := "2024-01-15 *\n  Assets:Bank:Current -> Expenses:Groceries \"Tesco \u2014 Weekly groceries\" 45.50 GBP\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestConvertBothExplicit(t *testing.T) {
	src := []byte(`2024-11-01 * "Transfer"
  Expenses:Food   100.00 GBP
  Assets:Bank:Current  -100.00 GBP
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	want := "2024-11-01 *\n  Assets:Bank:Current -> Expenses:Food \"Transfer\" 100.00 GBP\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestConvertPayeeNarration(t *testing.T) {
	src := []byte(`2024-01-01 * "Payee Co" "Some narration"
  Expenses:Food   50.00 GBP
  Assets:Bank
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"Payee Co — Some narration"`) {
		t.Errorf("expected combined payee+narration, got:\n%s", got)
	}
}

func TestConvertFlagged(t *testing.T) {
	src := []byte(`2024-01-01 ! "Flagged txn"
  Expenses:Food   25.00 GBP
  Assets:Bank
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), "2024-01-01 !\n") {
		t.Errorf("expected ! flag, got:\n%s", got)
	}
}

func TestConvertDropsDirectives(t *testing.T) {
	src := []byte(`2024-01-01 open Assets:Bank GBP
2024-01-01 balance Assets:Bank 1000.00 GBP
2024-01-01 * "Test"
  Expenses:Food   10.00 GBP
  Assets:Bank
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.Contains(s, "; DROPPED: open") {
		t.Error("expected DROPPED comment for open")
	}
	if !strings.Contains(s, "; DROPPED: balance") {
		t.Error("expected DROPPED comment for balance")
	}
	if !strings.Contains(s, "Assets:Bank -> Expenses:Food") {
		t.Error("expected converted transaction")
	}
}

func TestConvertSkipsNPostings(t *testing.T) {
	src := []byte(`2024-01-01 * "Split"
  Expenses:Food   30.00 GBP
  Expenses:Drinks   15.00 GBP
  Assets:Bank  -45.00 GBP
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "; SKIPPED (3 postings): Split") {
		t.Errorf("expected SKIPPED comment, got:\n%s", got)
	}
}

func TestConvertSimpleFile(t *testing.T) {
	src := []byte(`;  Test file
option "title" "Test"
2024-01-01 open Assets:Bank GBP
2024-01-01 * "Opening"
  Assets:Bank   1000.00 GBP
  Equity:Opening
`)
	got, err := Convert(src)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	// Should have dropped option and open
	if !strings.Contains(s, "; DROPPED: option") {
		t.Error("expected DROPPED option")
	}
	if !strings.Contains(s, "; DROPPED: open") {
		t.Error("expected DROPPED open")
	}
	// Should have converted the transaction
	if !strings.Contains(s, "Equity:Opening -> Assets:Bank") {
		t.Errorf("expected converted transaction, got:\n%s", got)
	}
}
