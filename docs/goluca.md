# Goluca

Movement-based accounting format using arrow notation to show flows between accounts.

Goluca uses *movements* instead of traditional postings, inspired by Pacioli's
Credit/Debit notation. Each movement transfers an amount between two accounts
using arrow operators (`->`, `//`, `>`), with linked movements grouped via the
`+` prefix.

See also: [ABNF and Plain Text Accounting](https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/),
[ABNF syntax comparison](https://www.bytestone.uk/posts/abnf/),
and [ABNF Standards and Extensions](abnf-variants.html) for details on
the non-standard constructs used in the grammars below.

## Journals and Source Files

The *books of account* are the totality of an entity's accounting
records. Within them, the *journal* is the chronological record of
movements — it is the source of truth and, in a modern PTA system, has
a plain text format (the source files). The *ledger* is the same data
reorganised by account, holding only one side of each movement; each
journal entry produces two ledger entries. The ledger is derived from
the journal, not stored independently.

A journal covers a single accounting period (e.g. a financial year)
and may be stored across one or more source files.

### Period linkage

Successive journals need a mechanism to link together so the full
history forms an auditable chain. Pacioli numbered his journals
sequentially, marking the first with a cross (✠). Goluca could adopt a
similar scheme — an explicit period declaration with a sequence number
and a reference to the prior period.

### Integrity via Merkle trees

To guarantee that a collection of journals has not been tampered with,
each period's closing state could include a cryptographic hash of its
contents. Chaining these hashes (as in a Merkle tree) would let anyone
verify the integrity of the entire ledger history from a single root
hash.

### Open questions

- Syntax for declaring a period and its sequence number.
- Whether the hash covers the raw source text or a canonical
  serialisation.
- How opening balances reference the prior period's closing hash.
- Whether sub-files within a period (e.g. monthly splits) also
  participate in the hash chain.

Details to be worked out.

## ABNF Grammar (proposed)

This extends the auto-generated grammar from `grammar.js` via
[tree-sitter2abnf](https://github.com/drummonds/tree-sitter2abnf)
with new directives for commodities, accounts, customers, configuration,
and full datetime support.
See [Goluca Datetime Formats](goluca-datetime.html) for details on
the datetime design and its relationship to ISO 8601 / RFC 3339.

```abnf
; @grammar "goluca"
; @extras (@pattern("\\r"))

source_file = *((directive / transaction / comment / %s"\n"))

; --- Directives ---

directive = commodity_directive / open_directive / option_directive / alias_directive / customer_directive / data_point

commodity_directive = [datetime _sp] %s"commodity" _sp commodity %s"\n" *metadata_line

open_directive = datetime _sp %s"open" _sp account [_sp commodity_list] %s"\n" *metadata_line

option_directive = %s"option" _sp option_key _sp option_value %s"\n"

alias_directive = %s"alias" _sp alias_name _sp account %s"\n"

customer_directive = %s"customer" _sp customer_name %s"\n" 1*customer_property

; --- Transactions ---

transaction = header 1*movement

header = datetime [knowledge_datetime] _sp flag [_sp payee] %s"\n"

movement = _sp [linked_prefix] @field(from) account _sp arrow _sp @field(to) account [_sp description] _sp amount _sp commodity %s"\n"

; --- Metadata ---

metadata_line = _indent metadata_key %s":" _sp metadata_value %s"\n"

customer_property = _indent (customer_account / customer_constraint / metadata_line)

customer_account = %s"account" _sp account %s"\n"

customer_constraint = %s"max-aggregate-balance" _sp amount _sp commodity %s"\n"

; --- Data Points (see goluca-parameters.html) ---

data_point = datetime [knowledge_datetime] %s"data" _sp param_name _sp param_value [_sp comment] %s"\n"

param_name = name_part *(%s":" name_part)

name_part = @pattern("[a-zA-Z][a-zA-Z0-9_-]*")

param_value = @pattern("[^\\n;]+")

; --- Tokens ---

comment = @token(@pattern("[#;]") @pattern("[^\\n]*"))

_sp = @pattern(" +")

_indent = @pattern("  ")

datetime = date [%s"T" time timezone]

date = @pattern("\\d{4}-\\d{2}-\\d{2}")

time = @pattern("\\d{2}:\\d{2}:\\d{2}") [fractional]

fractional = %s"." @pattern("\\d{1,6}")

timezone = %s"Z" / tz_offset

tz_offset = (%s"+" / %s"-") @pattern("\\d{2}:\\d{2}")

knowledge_datetime = %s"%" datetime

flag = (%s"*" / %s"!")

payee = @pattern("[^\\n]+")

linked_prefix = %s"+"

account = @pattern("[A-Za-z0-9][a-zA-Z0-9]*(:[A-Za-z0-9][a-zA-Z0-9-]*)+")

arrow = (%s"->" / %s"//" / %s"→" / %s">")

description = @pattern("\"[^\"]*\"")

amount = @pattern("-?[0-9][0-9,]*(\\.[0-9]+)?")

commodity = @token(@prec(1) @pattern("[A-Z][A-Z]+"))

commodity_list = commodity *(%s"," commodity)

option_key = @pattern("[a-z][a-z-]*")

option_value = @pattern("[^\\n]+")

alias_name = @pattern("[A-Za-z][a-zA-Z0-9-]*")

customer_name = @pattern("\"[^\"]*\"")

metadata_key = @pattern("[a-z][a-z-]*")

metadata_value = @pattern("[^\\n]+")
```

## ABNF Grammar (current goluca)

Auto-generated from the current
[tree-sitter-goluca](https://github.com/drummonds/tree-sitter-goluca)
`grammar.js` via
[tree-sitter2abnf](https://github.com/drummonds/tree-sitter2abnf).

<!-- GENERATED:GOLUCA_ABNF -->

<!-- GENERATED:GOLUCA_ROUNDTRIP -->

## ABNF Grammar (bytestone.uk)

Hand-written ABNF from the
[bytestone.uk article](https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/).
This predates the tree-sitter grammar and uses standard ABNF conventions
(WSP, CRLF, DQUOTE).

```abnf
journal = *(movement / linked-movement)

movement = date WSP flag CRLF WSP WSP from-account WSP arrow WSP
           to-account [WSP description] WSP amount WSP commodity CRLF

linked-movement = movement 1*("+" from-account WSP arrow WSP to-account
                  [WSP description] WSP amount [WSP commodity] CRLF)

date = 4DIGIT "-" 2DIGIT "-" 2DIGIT

flag = "*" / "!"

arrow = ">" / "->" / "//"

from-account = account

to-account = account

account = account-type *(":" name)

account-type = "Assets" / "Liabilities" / "Equity" / "Income" / "Expenses"

name = 1*VCHAR

description = DQUOTE *VCHAR DQUOTE

amount = 1*DIGIT ["." 1*DIGIT]

commodity = 1*ALPHA
```

## Characteristics

| Feature | Detail |
|---------|--------|
| Direction | Explicit — every movement names `from -> to` |
| Balancing | No implicit balancing; each movement is self-contained |
| Linked movements | `+` prefix groups multiple movements under one header |
| Arrow operators | `->` (transfer), `//` (split/fee), `>` (shorthand) |
| Accounts | Colon-separated hierarchy, any root name — see [accounts](goluca-accounts.html) |
| Commodities | Uppercase alpha, 2+ chars; must be declared with scaling — see [accounts](goluca-accounts.html#commodities) |
| Dates | ISO 8601 (`YYYY-MM-DD`) |
| Flags | `*` (cleared) or `!` (pending) |
| Comments | Lines starting with `#` or `;` |
| Description | Optional, double-quoted string on a movement line |
| Knowledge date | Optional `%YYYY-MM-DD` suffix on transaction date |
| Commodity declarations | `commodity GBP` or `2024-01-01 commodity GBP` with optional metadata |
| Account open | `2024-01-01 open Account:Name GBP` with optional metadata |
| Options | `option key value` for file-level configuration |
| Aliases | `alias ShortName Full:Account:Path` for shorthand references |
| Customers | `customer "Name"` with account links and constraints |
| Metadata | Indented `key: value` lines on directives and open statements |
| Data points | Time-stamped named parameters — see [parameters](goluca-parameters.html) |
| Hierarchical params | Metadata on parent `open` cascades to child accounts |

### Example

```goluca
2024-01-15 * Tesco
  Assets:Bank:Current -> Expenses:Groceries "Weekly groceries" 45.50 GBP
```

### Linked Movement Example

```goluca
2024-01-20 * Transfer with fee
  Assets:Bank:Current -> Assets:Savings "Monthly savings" 500.00 GBP
  +Assets:Bank:Current -> Expenses:BankCharges "Transfer fee" 1.50 GBP
```

### Knowledge Date Example

```goluca
; Transaction occurred Jan 15, booked on Jan 20 (e.g. credit card statement arrived late)
2024-01-15%2024-01-20 * Tesco
  Assets:CreditCard -> Expenses:Groceries "Weekly groceries" 45.50 GBP
```

### Commodity Declaration Examples

```goluca
commodity GBP                          ; undated, no metadata

2024-01-01 commodity GBP               ; dated, no metadata

2024-01-01 commodity GBP               ; dated, with metadata
  name: "British Pound Sterling"
  precision: 2

2024-01-01 commodity BTC
  name: "Bitcoin"
  precision: 8
```

### Account Open Examples

```goluca
2024-01-01 open Assets:Bank:Current GBP

2024-01-01 open Assets:Bank:Savings GBP,USD
  description: "ISA savings account"

; Hierarchical: parent sets defaults, children inherit
2024-01-01 open Assets:Bank
  currency: GBP
  institution: "Barclays"

2024-01-01 open Assets:Bank:Current
  ; inherits currency: GBP, institution: "Barclays"

2024-01-01 open Assets:Bank:Savings
  currency: GBP,USD   ; overrides parent's currency
```

### Option Directive Examples

```goluca
option operating-currency GBP
option require-accounts true
option title "My Personal Ledger"
```

### Alias Examples

```goluca
alias Groceries Expenses:Food:Groceries
alias Rent Expenses:Housing:Rent
alias Current Assets:Bank:Current

2024-01-15 * Tesco
  Current -> Groceries "Weekly shop" 45.50 GBP
```

### Customer Model Examples

```goluca
customer "John Smith"
  account Assets:Receivables:JohnSmith
  max-aggregate-balance 10000 GBP
  email: "john@example.com"
  payment-terms: "net-30"

customer "Acme Corp"
  account Assets:Receivables:AcmeCorp
  max-aggregate-balance 50000 GBP
  vat-number: "GB123456789"

2024-03-01 * Invoice 001
  Income:Consulting -> Assets:Receivables:JohnSmith "March consulting" 2500.00 GBP
```

### Hierarchical Parameter Inheritance

```goluca
2024-01-01 open Assets:Bank
  currency: GBP
  institution: "Barclays"

2024-01-01 open Assets:Bank:Current
  ; inherits currency: GBP, institution: "Barclays"

2024-01-01 open Assets:Bank:Savings
  currency: GBP,USD
  ; inherits institution: "Barclays", overrides currency
```

### Complete File Example

```goluca
; --- Options ---
option operating-currency GBP
option require-accounts true
option title "My Personal Ledger"

; --- Commodities ---
2024-01-01 commodity GBP
  name: "British Pound Sterling"
  precision: 2

2024-01-01 commodity USD
  name: "US Dollar"
  precision: 2

; --- Accounts ---
2024-01-01 open Assets:Bank
  currency: GBP
  institution: "Barclays"

2024-01-01 open Assets:Bank:Current
2024-01-01 open Assets:Bank:Savings
2024-01-01 open Assets:CreditCard
2024-01-01 open Expenses:Groceries
2024-01-01 open Expenses:BankCharges
2024-01-01 open Expenses:Housing:Rent
2024-01-01 open Income:Consulting

; --- Aliases ---
alias Groceries Expenses:Groceries
alias Current Assets:Bank:Current

; --- Customers ---
customer "John Smith"
  account Assets:Receivables:JohnSmith
  max-aggregate-balance 10000 GBP
  payment-terms: "net-30"

; --- Transactions ---
2024-01-15 * Tesco
  Current -> Groceries "Weekly groceries" 45.50 GBP

; Knowledge date: occurred Jan 18, booked Jan 22
2024-01-18%2024-01-22 * Online purchase
  Assets:CreditCard -> Expenses:Groceries "Delivery order" 32.00 GBP

2024-01-20 * Transfer with fee
  Assets:Bank:Current -> Assets:Bank:Savings "Monthly savings" 500.00 GBP
  +Assets:Bank:Current -> Expenses:BankCharges "Transfer fee" 1.50 GBP

2024-03-01 * Invoice 001
  Income:Consulting -> Assets:Receivables:JohnSmith "March consulting" 2500.00 GBP
```

## Semantic Rules

### require-accounts

When `option require-accounts true` is set:

- Every account used in a movement must have a prior `open` directive.
- Aliases are resolved **before** account validation — alias names themselves
  do not need an `open`, but their target accounts do.
- Commodity references must have prior `commodity` directives.
- Violations produce errors at parse/check time.

### Hierarchical parameter inheritance

- Metadata on an `open` directive cascades to all descendant accounts.
- A child `open` can override any inherited key.
- Only absent keys are inherited — from the nearest ancestor.
- Inheritance follows the colon-separated account hierarchy, not file order.
- The commodity list on the `open` line constrains movement commodities;
  `currency` metadata is informational only.

### Customer constraints

- `max-aggregate-balance` is checked against the aggregate balance of the
  customer's linked account.
- Implementations should report a **warning** (not error) when exceeded,
  since violations may be intentional during reconciliation.

## Differences: tree-sitter vs bytestone ABNF

The two grammars describe the same format but differ in detail:

| Aspect | tree-sitter2abnf | bytestone.uk |
|--------|-----------------|--------------|
| Top-level rule | `source_file` with transactions, comments, newlines | `journal` with movements only |
| Transaction structure | `header` + `1*movement` | Flat `movement` / `linked-movement` |
| Comments | Included as a grammar rule | Not defined |
| Arrow operators | `->`, `//`, `→` (Unicode), `>` | `->`, `//`, `>` |
| Account pattern | Regex-based (any uppercase start + colon-separated parts) | Enumerated top-level types (`Assets`, `Liabilities`, etc.) |
| Commodity | Uppercase, 2+ chars, with precedence | Any `1*ALPHA` |
| Amounts | Allows commas and optional decimals | Digits with optional decimals only |

**Note:** The proposed grammar extends the tree-sitter2abnf grammar with
directives (commodity, open, option, alias, customer), knowledge dates, and
metadata. These extensions are not yet implemented in the tree-sitter parser.
