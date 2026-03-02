# Beancount to go-luca Conversion

## go-luca ABNF Summary

```abnf
movement        = date SP flag LF indent from-account SP arrow SP to-account SP description SP amount SP commodity LF
linked-movement = date SP flag LF 1*(indent movement-line LF)
movement-line   = "+" SP from-account SP arrow SP to-account SP description SP amount SP commodity
date            = 4DIGIT "-" 2DIGIT "-" 2DIGIT
flag            = "*" / "!"
arrow           = "->"
account         = segment *(":" segment)
amount          = 1*DIGIT ["." 1*DIGIT]
commodity       = 1*ALPHA
description     = DQUOTE *UTF8 DQUOTE
```

## Mapping Table

| Beancount | go-luca | Notes |
|-----------|---------|-------|
| `transaction` (2 postings) | `movement` | Phase 1 |
| `transaction` (N postings) | `linked-movement` | Phase 2 (not yet implemented) |
| `*` flag | `*` | Direct mapping |
| `!` flag | `!` | Direct mapping |
| `"payee" "narration"` | `"payee — narration"` | Combined into description |
| `"narration"` | `"narration"` | Used as description |
| Account names | Account names | Direct mapping |
| Amount + commodity | amount commodity | Direct mapping |
| `open`, `close` | Dropped | No account lifecycle |
| `balance`, `pad` | Dropped | No assertions |
| `note`, `document`, `event`, `query`, `custom` | Dropped | No metadata directives |
| `option`, `include`, `plugin` | Dropped | No file-level config |
| `pushtag`/`poptag`, `pushmeta`/`popmeta` | Dropped | No scoped metadata |
| `commodity` | Dropped | No declarations |
| `price` | Dropped | No price directives |
| Tags `#tag`, links `^link` | Dropped | No equivalent |
| Posting metadata | Dropped | No equivalent |
| Cost `{...}`, price `@ ...` | Dropped | No equivalent |
| Comments | Dropped | No comment syntax in ABNF |

## Phase 1: 2-Posting Transaction Conversion

**From/to determination:**
1. Posting with positive explicit amount = **to** (receiving account)
2. Posting with no amount or negative amount = **from** (source account)
3. Both positive (edge case): first posting = to, second = from

### Example: Auto-balanced

```beancount
2024-01-15 * "Tesco" "Weekly groceries"
  Expenses:Groceries   45.50 GBP
  Assets:Bank:Current
```

```goluca
2024-01-15 *
  Assets:Bank:Current -> Expenses:Groceries "Tesco — Weekly groceries" 45.50 GBP
```

### Example: Both explicit

```beancount
2024-11-01 * "Complex Corp" "Full-featured transaction"
  Expenses:Food   100.00 GBP
  Assets:Bank:Current  -100.00 GBP
```

```goluca
2024-11-01 *
  Assets:Bank:Current -> Expenses:Food "Complex Corp — Full-featured transaction" 100.00 GBP
```

## Phase 2: N-Posting Transactions (Future)

Multi-posting transactions map to `linked-movement` with `+` continuation lines:

```beancount
2024-01-15 * "Split dinner"
  Expenses:Food        30.00 GBP
  Expenses:Drinks      15.00 GBP
  Assets:Bank:Current -45.00 GBP
```

```goluca
2024-01-15 *
  + Assets:Bank:Current -> Expenses:Food "Split dinner" 30.00 GBP
  + Assets:Bank:Current -> Expenses:Drinks "Split dinner" 15.00 GBP
```

## Edge Cases

- **Auto-balanced postings**: Missing amount inferred from the other posting's amount
- **Negative amounts**: `-100.00 GBP` indicates the from-account; absolute value used in output
- **Arithmetic amounts**: `100.00 + 50.00 GBP` — not supported in Phase 1, transaction skipped if encountered
