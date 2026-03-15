# Goluca Parameters (Data Points)

Parameters are time-stamped named values attached to the accounting
data. They record configuration, rates, limits, and other non-movement
facts that change over time.

## Syntax

```abnf
data_point = datetime [knowledge_datetime] %s"data" _sp param_name _sp param_value [_sp comment] %s"\n"

param_name  = name-part 1*(":" name-part)

name-part   = ALPHA *(ALPHA / DIGIT / "_" / "-")
```

### Examples

```goluca
2024-01-01 data gross_interest_rate 4.567%
2024-01-01 data balance:savings:my_account:gross_interest_rate 4.567%
2024-06-15 data balance:savings:my_account:gross_interest_rate 4.250%
```

## Hierarchical names

Parameter names use the same colon-separated hierarchy as accounts.
This gives a natural namespace and allows inheritance — a value set
at a parent level applies to all descendants unless overridden.

### Namespace structure

```
global_parameter
balance:parameter
balance:product:parameter
balance:product:account:parameter
```

| Level | Example | Scope |
|-------|---------|-------|
| Global | `gross_interest_rate` | All accounts |
| Balance sheet | `balance:gross_interest_rate` | All balance sheet items |
| Product | `balance:savings:gross_interest_rate` | All savings accounts |
| Account | `balance:savings:my_account:gross_interest_rate` | One account |

A more specific parameter overrides a less specific one. When looking
up `gross_interest_rate` for `balance:savings:my_account`, the system
checks (most specific first):

1. `balance:savings:my_account:gross_interest_rate`
2. `balance:savings:gross_interest_rate`
3. `balance:gross_interest_rate`
4. `gross_interest_rate`

### Relationship to account hierarchy

Parameter names do **not** follow the five Anglo-Saxon account types
(Assets, Liabilities, etc.). Like the French Plan Comptable Général,
parameters use a functional hierarchy where `balance` groups all
balance-sheet items regardless of whether a particular account is
currently an asset or liability.

This matters because some accounts change sign depending on activity —
a current account is normally an asset but becomes a liability when
overdrawn. The parameter hierarchy should be stable regardless of the
account's current balance:

```goluca
; The overdraft rate applies whether the account is in credit or debit
2024-01-01 data balance:current:overdraft_rate 19.9%

; The interest rate applies to credit balances
2024-01-01 data balance:current:credit_interest_rate 0.5%
```

## Value types

The bytestone.uk ABNF defines `value` as a bare string. In Go, untyped
string values create problems: every consumer must parse and validate
independently, errors surface late, and there is no way to enforce
units or precision at the type level.

### Proposed type system

Values should carry an explicit or inferred type. Options:

**Option A: Explicit type suffix**

```goluca
2024-01-01 data balance:savings:gross_interest_rate 4.567% rate
2024-01-01 data balance:savings:max_balance 85000.00 GBP amount
2024-01-01 data balance:savings:account_holder "Jane Smith" string
2024-01-01 data balance:savings:active true bool
```

**Option B: Inferred from value syntax**

| Pattern | Inferred type | Go type |
|---------|--------------|---------|
| `4.567%` | Rate (percentage) | `decimal` or `int64` (basis points) |
| `85000.00 GBP` | Amount with commodity | `Amount` (value + commodity) |
| `"Jane Smith"` | String | `string` |
| `true` / `false` | Boolean | `bool` |
| `2024-01-01` | Date | `time.Time` |
| `30` | Integer | `int64` |
| `P30D` | Duration (ISO 8601) | `time.Duration` |

**Option C: Schema declaration**

Define parameter types in a separate schema, then values are validated
against it:

```goluca
option param-type gross_interest_rate rate
option param-type max_balance amount
option param-type account_holder string
```

### Rate representation

Rates deserve special attention. A percentage like `4.567%` could be
stored as:

| Representation | Value | Precision | Notes |
|----------------|-------|-----------|-------|
| Float | 0.04567 | ~15 sig figs | Floating-point errors accumulate |
| Basis points (int) | 456.7 | 0.1 bp | Not exact for all rates |
| Millionths (int) | 45670 | 0.0001% | Exact for 4-digit rates |
| String | "4.567%" | Arbitrary | Must parse on every use |

For Go, an integer representation (basis points × 10 or millionths)
avoids floating-point issues while keeping arithmetic fast.

### Amount values

When a parameter value includes a commodity (e.g. `85000.00 GBP`),
the precision and scaling rules from the commodity declaration apply.
This ties parameters back to the commodity system — the value
`85000.00` in `GBP` means 8,500,000 pence internally.

## ABNF (proposed)

Adding data points to the proposed goluca grammar:

```abnf
; --- added to directive ---
directive = commodity_directive / open_directive / option_directive
          / alias_directive / customer_directive / data_point

data_point = datetime [knowledge_datetime] %s"data" _sp param_name _sp param_value [_sp comment] %s"\n"

param_name = name-part *(":" name-part)

param_value = amount_value / rate_value / bool_value / date / duration / quoted_string / bare_string

amount_value = amount _sp commodity

rate_value = @pattern("-?[0-9]+(\\.[0-9]+)?%")

bool_value = %s"true" / %s"false"

duration = @pattern("P(\\d+Y)?(\\d+M)?(\\d+D)?(T(\\d+H)?(\\d+M)?(\\d+S)?)?")

quoted_string = @pattern("\"[^\"]*\"")

bare_string = @pattern("[^ \\t\\n;][^\\n;]*")
```

## Open questions

- Whether to use explicit type suffixes, inference, or schema
  declarations (or some combination).
- How rate precision interacts with the commodity scaling system.
- Whether parameters should support multi-line values or structured
  data (e.g. schedules of rates by band).
- Interaction with the period/journal model — do parameters have
  journal scope or ledger scope?
- Whether parameter history should be queryable as a time series
  (e.g. "what was the interest rate on 2024-06-01?").
