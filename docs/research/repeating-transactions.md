# Repeating Transactions in Plain Text Accounting

Research into how PTA formats handle repeating transactions, and how
this intersects with human readability and information density.

## Motivation

When creating test banking products it is natural to say "repeat this
exact movement 28 times" rather than writing out 28 identical
transactions with incrementing dates. This is a count-based repetition
— distinct from calendar-driven recurrence ("every month") that most
PTA tools focus on.

The question: how should a PTA format express repetition, and what are
the trade-offs between compactness and explicitness?

## Survey of Existing Formats

### Ledger-cli

Ledger has two related features:

**Periodic transactions** use a `~` prefix with a period expression.
These define budget amounts, not real transactions:

```ledger
~ Monthly
  Expenses:Rent         $2000.00
  Assets:Bank:Checking
```

**Automated transactions** use a `=` prefix and pattern-match against
other transactions to add postings:

```ledger
= /Groceries/
  Budget:Food          -1.0
  Budget:Food:Unspent
```

Neither generates actual journal entries. Ledger is a reporting tool —
it never modifies the journal file. External tools like
[ledgerbil](https://github.com/scarpent/ledgerbil) and `ledger-recur`
exist to materialise recurring entries into the file.

**Repetition model:** calendar-interval only, no count-based repetition.
Generated at report time, not stored.

### hledger

hledger extends Ledger's periodic transactions with a `--forecast` flag
that generates real transactions into the report:

```hledger
~ every 2 weeks from 2024/06 to 2024/09
  Assets:Bank:Checking   $1500.00
  Income:Salary
```

Period expressions are rich:

- `~ monthly`
- `~ every 2 weeks from 2024/01 to 2024/06`
- `~ every 2nd Thursday of month from 2024/01 to 2024/04`
- `~ every nov 29th from 2024 to 2026`

Generated transactions carry a `recur` tag with the period expression
as value, so downstream tools can distinguish them from manually
entered transactions.

The [hledger-forecast](https://github.com/olimorris/hledger-forecast)
plugin adds YAML-based forecast definitions with more control.

**Repetition model:** calendar-interval with date ranges, no explicit
count. Transactions are generated at query time via `--forecast`, not
stored in the journal.

### Beancount

Beancount has **no built-in recurring transaction support**. The
philosophy is that the journal contains only actual, occurred
transactions. Repetition is handled by plugins:

**beancount-periodic** — uses metadata to drive generation:

```beancount
2024-03-31 * "Provider" "Net Fee"
  recur: "1 Year /Monthly"
  Liabilities:CreditCard  -50 USD
  Expenses:CommunicationFee  50 USD
```

**beancount-repete** — natural language scheduling:

```beancount
2024-01-01 ! "Supermarket" "Weekly shop"
  repete: "weekly on Tuesday until March 2024"
  Assets:Bank  -75.00 GBP
  Expenses:Groceries  75.00 GBP
```

**Fava forecast plugin** — tag-based syntax in narration:

```beancount
2024-01-01 # "Rent payment [MONTHLY]"
  Expenses:Housing:Rent   2500.00 USD
  Assets:Checking        -2500.00 USD
```

Can use `REPEAT n TIMES` for count-based repetition — the only PTA
ecosystem that supports this natively (via plugin).

**Repetition model:** plugin-driven, varies by plugin. Most are
calendar-interval; Fava's plugin supports count-based. Expanded at
load time by the plugin.

### GnuCash

GnuCash (not plain-text, but relevant as a reference) has full
scheduled transaction support in the GUI:

- Frequency (weekly, monthly, etc.)
- Start/end date
- Number of occurrences (count-based)
- "Since Last Run" assistant that creates transactions on startup

**Repetition model:** calendar + count, stored as schedule metadata
separately from the ledger. Materialised into the register
interactively.

### Coin (mkobetic/coin)

Go implementation of the Ledger model. No recurring transaction
support found — focuses on the core parse/report cycle.

### Transity

YAML-based format that models flows rather than postings (similar
philosophy to Goluca). No recurring transaction support found in the
format itself.

### Knut

Go-based PTA tool. Supports prices, transactions, balance assertions,
and value directives. No recurring transaction syntax found.

## Summary Table

| Tool | Built-in? | Calendar interval | Count-based | Storage |
|------|-----------|-------------------|-------------|---------|
| Ledger | `~` periodic | Yes | No | Report-time only |
| hledger | `~` periodic + `--forecast` | Yes (rich expressions) | No | Generated at query time |
| Beancount | No (plugins) | Yes (via plugins) | Yes (Fava: `REPEAT n TIMES`) | Expanded by plugin at load |
| GnuCash | Yes (GUI) | Yes | Yes (occurrence count) | Schedule metadata + materialise |
| Coin | No | No | No | N/A |
| Transity | No | No | No | N/A |
| Knut | No | No | No | N/A |

## The Readability / Redundancy Spectrum

PTA formats sit on a spectrum between two extremes:

### Fully expanded (maximum redundancy)

Every transaction is written out explicitly. 28 monthly interest
payments means 28 transaction blocks in the file.

```goluca
2024-01-15 * Monthly interest
  Assets:Bank:Current -> Income:Interest "Jan interest" 10.00 GBP

2024-02-15 * Monthly interest
  Assets:Bank:Current -> Income:Interest "Feb interest" 10.00 GBP

; ... 26 more ...
```

**Pros:**
- Maximally readable — no mental model required beyond "read the file"
- Each entry is independently verifiable
- `grep`, `diff`, `git blame` all work naturally
- No parser support needed for repetition
- Clear audit trail — every transaction is visible

**Cons:**
- Verbose — obscures the pattern behind repetitive noise
- Error-prone to maintain — changing the amount means 28 edits
- Violates DRY — the *intent* ("28 identical payments") is lost
- The file is dominated by mechanical repetition, not information

### Fully compact (maximum information density)

A single directive generates all 28 transactions:

```
repeat 28 monthly from 2024-01-15
  Assets:Bank:Current -> Income:Interest "Monthly interest" 10.00 GBP
```

**Pros:**
- Intent is explicit — "28 monthly payments" is right there
- Single point of change
- File contains only decisions, not their mechanical expansion
- Compact test data generation

**Cons:**
- Requires tool support to expand
- Not self-evident — you need to understand the `repeat` directive
- Raw file doesn't show actual transaction dates
- Harder to verify individual transactions
- Audit trail is less clear — was the 15th payment actually made?

### The middle ground

Most PTA tools sit closer to the expanded end. The journal is the
*book of record* — it should contain what actually happened, not
templates for what might happen. Periodic/forecast features are
typically:

1. Separate from the main journal (hledger's `--forecast`)
2. Clearly marked as generated (hledger's `recur` tag)
3. Plugin territory, not core syntax (Beancount)

This reflects an accounting principle: the journal records *facts*.
Repetition rules are *policy* that generates facts.

## Design Considerations for Goluca

### Use case: test data generation vs journal entries

The motivating use case — "repeat this movement 28 times for a test
banking product" — is closer to test data generation than journal
recording. This distinction matters:

| Concern | Test data | Journal |
|---------|-----------|---------|
| Readability | Intent matters most | Individual entries matter most |
| Auditability | Not required | Essential |
| Compactness | High value | Lower priority |
| Tool support | Can require expansion step | Should work with basic tools |

### Proposed design: `~` flag with repeat clause

Use `~` as a third flag alongside `*` (cleared) and `!` (pending),
followed by a repeat clause specifying either a count or an end date,
and an interval. This echoes the Ledger/hledger convention where `~`
signals periodicity, but brings it into the transaction header as a
proper flag.

The existing Goluca grammar defines:

```abnf
header = datetime [knowledge_datetime] _sp flag [_sp payee] %s"\n"
flag   = (%s"*" / %s"!")
```

The extension adds `~` as a flag value and a repeat clause between the
flag and the optional payee:

```abnf
flag           = (%s"*" / %s"!" / %s"~")
repeat_clause  = %s"repeat" _sp (count_repeat / until_repeat)
count_repeat   = count _sp interval
until_repeat   = interval _sp %s"until" _sp datetime
count          = @pattern("[1-9][0-9]*")
interval       = %s"daily" / %s"weekly" / %s"monthly" / %s"yearly"
```

The header becomes:

```abnf
header = datetime [knowledge_datetime] _sp flag [_sp repeat_clause] [_sp payee] %s"\n"
```

The repeat clause is only valid when the flag is `~`.

#### Syntax examples

**Count-based** — repeat 26 times at daily intervals:

```goluca
2024-01-15 ~ repeat 26 daily Interest accrual
  Assets:Bank:Current -> Income:Interest "Daily interest" 0.38 GBP
```

Expands to 26 transactions: 2024-01-15, 2024-01-16, ..., 2024-02-09.

**Until-based** — repeat monthly until an end date:

```goluca
2024-01-15 ~ repeat monthly until 2024-12-15 Monthly interest
  Assets:Bank:Current -> Income:Interest "Monthly interest" 10.00 GBP
```

Expands to 12 transactions: 2024-01-15, 2024-02-15, ..., 2024-12-15.

**Count + monthly** — exactly 28 monthly payments:

```goluca
2024-01-01 ~ repeat 28 monthly Mortgage payment
  Assets:Bank:Current -> Liabilities:Mortgage "Monthly repayment" 850.00 GBP
```

#### Why `~` as a flag

| Aspect | Rationale |
|--------|-----------|
| Familiar | Ledger/hledger use `~` for periodic transactions |
| Distinct | Visually different from `*` and `!` — stands out in the file |
| Parseable | Single character, easy to match in tree-sitter |
| Semantic | Signals "this is a template, not a single event" |

#### Interval semantics

| Interval | Meaning |
|----------|---------|
| `daily` | +1 calendar day |
| `weekly` | +7 calendar days |
| `monthly` | Same day-of-month in next month. If the day doesn't exist (e.g. 31st in February), roll back to last day of month |
| `yearly` | Same day-of-year. Feb 29 in non-leap years becomes Feb 28 |

#### Edge cases

**Day-of-month overflow:** `2024-01-31 ~ repeat 3 monthly` produces
2024-01-31, 2024-02-29 (leap year), 2024-03-31. The rule is: clamp
to last day of month, then resume the original day when possible.

**Until date not on boundary:** `2024-01-15 ~ repeat daily until 2024-01-17`
produces 3 transactions (15th, 16th, 17th inclusive).

**Knowledge dates:** Each expanded transaction gets its own date but
the knowledge date (if present) applies only to the template — it
records when the *repetition rule* was known, not when each instance
was known.

#### Interaction with linked movements

Linked movements (`+` prefix) repeat as a group:

```goluca
2024-01-20 ~ repeat 12 monthly Transfer with fee
  Assets:Bank:Current -> Assets:Savings "Monthly savings" 500.00 GBP
  +Assets:Bank:Current -> Expenses:BankCharges "Transfer fee" 1.50 GBP
```

Each expansion produces a complete transaction with both movements.

#### Expansion model

Two approaches, both viable:

1. **Expand at parse time** — the parser/tool generates N concrete
   transactions. The `~` transaction is never stored in a ledger;
   it's syntactic sugar. This is the simplest model and matches the
   test-data-generation use case.

2. **Expand on demand** — the `~` transaction is preserved in the
   AST and expanded only when generating reports or output. This
   keeps the source file compact but requires all downstream tools
   to understand `~`.

Recommendation: **expand at parse time** for the initial
implementation. A `goluca expand` subcommand could also materialise
the transactions into a file for auditing.

## Open Questions

1. **Additional intervals?** `fortnightly` (every 2 weeks)?
   `quarterly`? Or keep the set minimal and add later.

2. **Multiplied intervals?** Should `repeat 6 2-weekly` or
   `repeat 4 3-monthly` be supported, or is that over-engineering
   for now?

3. **Variable amounts?** The current design repeats the exact same
   movement. If amounts need to vary (e.g. amortisation schedules),
   that's a template language — a much larger design surface best
   deferred.

4. **Interaction with `require-accounts`?** A `~` transaction should
   validate accounts and commodities the same as `*` or `!`.

5. **Should expanded transactions carry provenance metadata?** e.g.
   `generated-from: line 42` or a `recur` tag like hledger. Useful
   for debugging but adds noise.

## References

- [Ledger periodic transactions](https://ledger-cli.org/doc/ledger3.html)
- [hledger forecasting](https://hledger.org/forecasting.html)
- [hledger budgeting and forecasting](https://hledger.org/budgeting-and-forecasting.html)
- [beancount-periodic](https://github.com/dallaslu/beancount-periodic)
- [beancount-repete](https://github.com/jpluscplusm/beancount-repete)
- [Fava forecast plugin](https://beancount.io/docs/Tips/forecast-plugin)
- [GnuCash scheduled transactions](https://wiki.gnucash.org/wiki/Scheduled_Transactions)
- [hledger-forecast](https://github.com/olimorris/hledger-forecast)
- [plaintextaccounting.org](https://plaintextaccounting.org/)
