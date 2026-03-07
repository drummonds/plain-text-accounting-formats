# Plain Text Accounting Formats

This project compares three plain text accounting (PTA) formats, each with a
[tree-sitter](https://tree-sitter.github.io/) grammar for parsing and syntax highlighting.

- [README](README.html) — project overview, install, CLI and library usage
- [ROADMAP](ROADMAP.html) — what's done and what's next
- [Parser demo](demo.html) — side-by-side syntax-highlighted output from tree-sitter grammars

## The Formats

| Format | Style | Source of Truth |
|--------|-------|-----------------|
| [Beancount](internal/beancount.html) | Transaction with postings (debit/credit implicit) | [beancount.github.io](https://beancount.github.io/) |
| [Goluca](internal/goluca.html) | Directed movements with `->` arrows | [tree-sitter-goluca](https://github.com/drummonds/tree-sitter-goluca) |
| [PTA](internal/pta.html) | Directed movements (goluca-compatible, simplified) | [tree-sitter-pta](https://github.com/drummonds/tree-sitter-pta) |

## Key Differences

**Beancount** uses the traditional ledger model: a transaction header followed by
two or more posting lines, each naming an account and optionally an amount.
Amounts balance implicitly across postings.

**Goluca** (go-luca) replaces implicit balancing with explicit directional movements.
Each line names a `from-account -> to-account` pair with an amount, making the
flow of money visible. Inspired by Pacioli's double-entry notation.

**PTA** shares goluca's directional style but with a simplified grammar.
Both use the same arrow operators (`->`, `//`, `>`) and linked-movement syntax.

### Same Transaction in Three Formats

**Beancount:**
```beancount
2024-01-15 * "Tesco" "Weekly groceries"
  Expenses:Groceries   45.50 GBP
  Assets:Bank:Current
```

**Goluca:**
```goluca
2024-01-15 * Tesco
  Assets:Bank:Current -> Expenses:Groceries "Weekly groceries" 45.50 GBP
```

**PTA:**
```pta
2024-01-15 * Tesco
  Assets:Bank:Current -> Expenses:Groceries Weekly groceries 45.50 GBP
```

## ABNF Grammar Summaries

Each format has a formal grammar defined in ABNF. See the individual format pages
for full definitions. For background on ABNF in accounting contexts, see
[ABNF and Plain Text Accounting](https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/)
on bytestone.uk.

## Related

- [Conversion: Beancount to Goluca](internal/beancount-to-goluca.html) — mapping rules and examples
- [mkobetic/coin](internal/coin.html) — a Go PTA tool with a different take on the ledger format
- [plaintextaccounting.org](https://plaintextaccounting.org/) — community resource covering all PTA tools
