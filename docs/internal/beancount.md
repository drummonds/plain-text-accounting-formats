# Beancount

[Beancount](https://beancount.github.io/) is a double-entry bookkeeping language
created by Martin Blais. It is the most widely-used plain text accounting format
after ledger-cli.

## Source of Truth

The canonical grammar is defined by the beancount project itself. The tree-sitter
grammar used in this project is
[tree-sitter-beancount](https://github.com/polarmutex/tree-sitter-beancount).

## ABNF Summary

```abnf
journal     = *entry
entry       = transaction / directive / comment
transaction = date SP txn [SP payee] SP narration LF 2*posting
posting     = indent account [SP amount] LF
amount      = number SP commodity
txn         = "*" / "!" / "txn"
date        = 4DIGIT "-" 2DIGIT "-" 2DIGIT
account     = segment *(":" segment)
segment     = ALPHA *ALPHANUM
commodity   = 1*UPPER
```

### Directives

Beancount supports many directive types beyond transactions:

| Directive | Purpose |
|-----------|---------|
| `open` / `close` | Account lifecycle |
| `balance` | Balance assertion |
| `pad` | Auto-pad to match balance |
| `note` / `document` | Attach metadata to accounts |
| `event` / `query` / `custom` | Reporting metadata |
| `price` / `commodity` | Market data and declarations |
| `option` / `include` / `plugin` | File-level configuration |
| `pushtag` / `poptag` | Scoped tagging |

## Example

```beancount
option "operating_currency" "GBP"

2024-01-01 open Assets:Bank:Current  GBP
2024-01-01 open Expenses:Groceries   GBP

2024-01-15 * "Tesco" "Weekly groceries"
  Expenses:Groceries   45.50 GBP
  Assets:Bank:Current

2024-01-31 balance Assets:Bank:Current  954.50 GBP
```

## Characteristics

- **Implicit balancing**: postings in a transaction must sum to zero; one posting may omit its amount
- **Rich metadata**: tags (`#tag`), links (`^link`), key-value metadata on postings
- **Cost/price tracking**: `{cost}` and `@ price` syntax for investment accounting
- **Strict account types**: accounts must start with `Assets`, `Liabilities`, `Equity`, `Income`, or `Expenses`
- **Plugin system**: extensible via Python plugins

## See Also

- [Conversion: Beancount to Goluca](beancount-to-goluca.html)
- [Format comparison](../index.html)
