# Goluca (go-luca)

Goluca is a plain text accounting format that uses directional movements
instead of traditional debit/credit postings. It is inspired by Luca Pacioli's
original double-entry bookkeeping notation.

## Source of Truth

- Grammar: [tree-sitter-goluca](https://github.com/drummonds/tree-sitter-goluca)
- Background: [Journal entries as plain text](https://www.bytestone.uk/afp/plain-text-accounting/journalasplaintext/) on bytestone.uk

## ABNF Summary

```abnf
journal         = *(movement / linked-movement / comment)
movement        = date SP flag [SP payee] LF
                  indent from-account SP arrow SP to-account
                  [SP description] SP amount SP commodity LF
linked-movement = date SP flag [SP payee] LF
                  1*(indent "+" SP from-account SP arrow SP to-account
                  [SP description] SP amount SP commodity LF)
date            = 4DIGIT "-" 2DIGIT "-" 2DIGIT
flag            = "*" / "!"
arrow           = "->" / "//" / ">" / U+2192
account         = segment *(":" segment)
amount          = ["-"] 1*DIGIT ["." 1*DIGIT]
commodity       = 1*UPPER
description     = DQUOTE *UTF8 DQUOTE
comment         = ("#" / ";") *UTF8
```

## Example

```goluca
# Opening balances
2024-01-01 * Opening
  Equity:OpeningBalances -> Assets:Bank:Current "Opening balance" 1000.00 GBP

# Simple movement
2024-01-15 * Tesco
  Assets:Bank:Current -> Expenses:Groceries "Weekly groceries" 45.50 GBP

# Linked movement (split transaction)
2024-01-20 * Split dinner
  + Assets:Bank:Current -> Expenses:Food "Food" 30.00 GBP
  + Assets:Bank:Current -> Expenses:Drinks "Drinks" 15.00 GBP
```

## Characteristics

- **Explicit direction**: every movement names both the source and destination account with an arrow (`->`)
- **No implicit balancing**: amounts are always explicit on every line
- **Linked movements**: multi-leg transactions use `+` continuation lines
- **Multiple arrow styles**: `->`, `//`, `>`, and Unicode `->` are all valid
- **Payee on header line**: payee is free text after the flag, not quoted
- **Description per leg**: each movement line can have its own quoted description
- **Comments**: `#` or `;` line comments

## Comparison with Beancount

| Aspect | Beancount | Goluca |
|--------|-----------|--------|
| Transaction model | Header + postings | Header + directed movements |
| Balancing | Implicit (postings sum to zero) | Explicit (each line has amount) |
| Direction | Implied by sign/account type | Explicit `from -> to` |
| Account declarations | Required (`open`/`close`) | Not used |
| Metadata directives | Many (`balance`, `note`, etc.) | None |
| Cost/price tracking | Yes (`{cost}`, `@ price`) | Not supported |

## See Also

- [Conversion: Beancount to Goluca](beancount-to-goluca.html)
- [Format comparison](../index.html)
