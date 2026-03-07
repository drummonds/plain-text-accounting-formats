# mkobetic/coin

[coin](https://github.com/mkobetic/coin) is a Go implementation of a plain text
accounting tool that follows the ledger-cli model. It is an interesting point of
comparison because it makes deliberate trade-offs that differ from both beancount
and the goluca/PTA approach.

## Design Philosophy

Coin is described by its author as "a personal exploration of the domain and a
purpose tailored implementation." It follows ledger-cli patterns but makes its
own choices:

- **No quoted commodity names**: coin sacrifices flexibility to avoid requiring
  quotation marks around commodity names
- **Multi-file ledgers**: designed for splitting ledger data across multiple files
- **Minimal dependencies**: pure Go with no external parser generators

## Format

Coin uses the traditional ledger format — transaction headers followed by posting
lines with accounts and amounts:

```ledger
2024-01-15 * Tesco - Weekly groceries
  Expenses:Groceries                45.50 GBP
  Assets:Bank:Current              -45.50 GBP
```

This is the same structural model as beancount (header + postings) but with
ledger-cli syntax rather than beancount syntax.

## Tooling

Coin provides a suite of CLI tools:

| Command | Purpose |
|---------|---------|
| `coin balance` | Balance report |
| `coin register` | Transaction register |
| `coin accounts` | List accounts |
| `coin commodities` | List commodities |
| `gc2coin` | Import from GnuCash |
| `ofx2coin` | Import from OFX/QFX bank files |
| `csv2coin` | Import from CSV |
| `coin2html` | Export to HTML |

## Comparison

| Aspect | Beancount | Coin | Goluca/PTA |
|--------|-----------|------|------------|
| Heritage | Own format | Ledger-cli | Original |
| Transaction model | Header + postings | Header + postings | Directed movements |
| Balancing | Implicit | Implicit | Explicit |
| Language | Python | Go | Go (tree-sitter) |
| Reporting | Built-in (fava) | Built-in CLI | Parsing only |
| Import tools | No | GnuCash, OFX, CSV | No |

## Why It's Interesting

Coin demonstrates that the ledger-cli model works well in Go without needing
a parser generator. Its import tools (`gc2coin`, `ofx2coin`, `csv2coin`) show
the practical side of PTA — getting real financial data into plain text format.

The goluca/PTA formats take a fundamentally different approach by making money
flow direction explicit, but coin shows there is still value in the traditional
posting model when paired with good tooling.

## Links

- [GitHub: mkobetic/coin](https://github.com/mkobetic/coin)
- [plaintextaccounting.org](https://plaintextaccounting.org/)
