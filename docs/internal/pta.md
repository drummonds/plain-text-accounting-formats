# PTA (Plain Text Accounting)

PTA is a simplified plain text accounting format sharing goluca's directional
movement model. It uses the same arrow operators and linked-movement syntax
but with a lighter grammar.

## Source of Truth

- Grammar: [tree-sitter-pta](https://github.com/drummonds/tree-sitter-pta)

## ABNF Summary

```abnf
journal         = *(transaction / comment)
transaction     = date SP flag [SP payee] LF
                  1*(indent posting-leg LF)
posting-leg     = ["+" SP] from-account SP arrow SP to-account
                  SP description SP amount SP commodity
date            = 4DIGIT "-" 2DIGIT "-" 2DIGIT
flag            = "*" / "!"
arrow           = "->" / "//" / ">" / U+2192
account         = segment *(":" segment)
amount          = ["-"] 1*DIGIT ["." 1*DIGIT]
commodity       = 1*UPPER
description     = *UTF8          ; unquoted free text
comment         = ("#" / ";") *UTF8
```

## Example

```pta
# Simple transaction
2024-01-15 * Tesco
  Assets:Bank:Current -> Expenses:Groceries Weekly groceries 45.50 GBP

# Linked transaction (split)
2024-01-20 * Split dinner
  + Assets:Bank:Current -> Expenses:Food Food 30.00 GBP
  + Assets:Bank:Current -> Expenses:Drinks Drinks 15.00 GBP
```

## Characteristics

- **Directed movements**: same `from -> to` model as goluca
- **Unquoted descriptions**: descriptions do not require quotes (unlike goluca)
- **Minimal syntax**: no account declarations, metadata, or directives
- **Comments**: `#` or `;` line comments
- **Same arrow operators**: `->`, `//`, `>`, Unicode arrow

## Comparison with Goluca

PTA and goluca are closely related. The key differences:

| Aspect | Goluca | PTA |
|--------|--------|-----|
| Descriptions | Quoted (`"text"`) | Unquoted (free text) |
| Grammar complexity | 51 parser states | 58 parser states |
| Payee handling | Free text on header | Free text on header |

Both formats are intentionally simple, focusing purely on recording movements
of money between accounts.

## See Also

- [Format comparison](../index.html)
