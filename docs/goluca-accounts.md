# Goluca Accounts and Commodities

## Accounts

An account is a colon-separated hierarchical name:

```
Assets:Bank:Current
Liabilities:CreditCard:Amex
Equity:Retained-Earnings
Income:Salary
Expenses:Food:Groceries
```

### Account name rules

```abnf
account      = account-root 1*(":" name-part)

account-root = 1*(ALPHA / DIGIT)

name-part    = ALPHA *(ALPHA / DIGIT / "-")
```

- The first component is the account root. It is not restricted to a
  fixed set of names — see *Beyond five account types* below.
- Subsequent components start with a letter and may contain letters,
  digits, and hyphens.

### Hierarchical balances

The hierarchy is implicit — `Assets:Bank:Current` makes `Assets:Bank`
a logical parent. The parent need not be declared, but if it is (via
`open`), its balance is the aggregate of all descendant accounts.

This means you can query at any level of the hierarchy:

| Query | Returns |
|-------|---------|
| `Assets:Bank:Current` | Balance of the current account only |
| `Assets:Bank` | Sum of `Current`, `Savings`, and any other children |
| `Assets` | Sum of all asset accounts |

### Beyond five account types

The Anglo-Saxon accounting tradition uses five root categories:
Assets, Liabilities, Equity, Income, and Expenses. Beancount hard-codes
these five. However, many national chart-of-accounts standards use
different structures:

**French Plan Comptable Général (PCG)** — eight numbered classes:

| Class | Name | Anglo-Saxon equivalent |
|-------|------|----------------------|
| 1 | Comptes de capitaux | Equity + long-term Liabilities |
| 2 | Comptes d'immobilisations | Fixed Assets |
| 3 | Comptes de stocks | Current Assets (inventory) |
| 4 | Comptes de tiers | Receivables + Payables |
| 5 | Comptes financiers | Cash + short-term Assets |
| 6 | Comptes de charges | Expenses |
| 7 | Comptes de produits | Income |
| 8 | Comptes spéciaux | Off-balance-sheet / special |

> See: [French generally accepted accounting principles](https://en.wikipedia.org/wiki/French_generally_accepted_accounting_principles)

**Swedish BAS chart of accounts** — also eight classes, numbered 1–8,
with a different split. Class 1 covers assets, class 2 equity and
liabilities, classes 3–4 income, and classes 5–8 expenses of various
kinds.

> See: [BAS chart of accounts](https://www.bas.se/english/chart-of-account/)

Goluca therefore does **not** restrict account roots to the five
Anglo-Saxon names. Any valid identifier may be used as the root. The
mapping from account roots to the fundamental accounting equation
(A = L + E) is a semantic concern handled by configuration, not by
the grammar.

### Aliases

Aliases provide shorthand names for frequently used accounts:

```goluca
alias Groceries Expenses:Food:Groceries
alias Current Assets:Bank:Current
alias Rent Expenses:Housing:Rent
```

After declaration, the alias can be used anywhere an account name is
expected:

```goluca
2024-01-15 * Tesco
  Current -> Groceries "Weekly shop" 45.50 GBP
```

Aliases are resolved before any validation — the alias name itself
does not need an `open` directive, but its target account does (when
`require-accounts` is enabled).

Aliases are particularly useful with numbered account systems (PCG,
BAS) where the raw account paths are opaque:

```goluca
alias Fournisseurs 4:Fournisseurs
alias VenteMarchandises 7:Ventes:Marchandises

2024-03-01 * Facture 042
  Fournisseurs -> VenteMarchandises "Vente mars" 1500.00 EUR
```

### Comparison with beancount

Beancount hard-codes the five Anglo-Saxon account types and enforces
them in the grammar. Its `open` directive optionally constrains which
commodities an account may hold. Goluca makes commodity declarations
mandatory (see below) and leaves account root names unconstrained.

> See: [Beancount Language Syntax — Accounts](https://beancount.github.io/docs/beancount_language_syntax.html)

## Commodities

A commodity is an uppercase identifier of two or more characters
(e.g. `GBP`, `USD`, `BTC`).

### Why commodity declarations are required

In beancount a commodity springs into existence on first use. Goluca
requires an explicit `commodity` directive because the *scaling* of a
commodity to its units must be defined — without it, arithmetic
precision is ambiguous.

```goluca
2024-01-01 commodity GBP
  name: "British Pound Sterling"
  precision: 2
  unit: "pence"
  scaling: 100
```

### Precision and scaling

The `precision` metadata on a commodity defines how many decimal
places are stored and displayed. The `scaling` metadata defines how
many minor units make one major unit. These are related but distinct:

| Commodity | Unit | Scaling | Precision | Notes |
|-----------|------|---------|-----------|-------|
| GBP | pence | 100 | 2 | Rounded to the nearest penny |
| CHF | centime | 100 | 2 | Rounded to the nearest 5 centimes |
| BTC | satoshi | 100,000,000 | 8 | |
| GBp | penny | 1 | 0 | Pence as the base unit |

### Precision conversion

When the same underlying currency is used at different precisions —
e.g. `GBP` (2 decimal places) vs an internal accrual commodity with
5 decimal places of pence — converting between them requires explicit
scaling. The conversion is deterministic (multiply or divide by a
power of 10) but the rounding direction must be specified.

For example, converting 1.23456 internal-pence to display pence:

| Rounding | Result | Use case |
|----------|--------|----------|
| Truncate | 1.23 | Conservative (favours payer) |
| Round half-up | 1.23 | Standard commercial rounding |
| Round half-even | 1.23 | Banker's rounding (minimises bias) |
| Ceiling | 1.24 | Conservative (favours payee) |

The rounding rule should be declared on the commodity or on the
conversion relationship between related commodities.

### Rounding

Not all commodities round to their smallest unit. Swiss francs, for
example, round to the nearest 5 centimes (0.05 CHF). The rounding
rule is a property of the commodity, not the precision:

```goluca
2024-01-01 commodity CHF
  name: "Swiss Franc"
  precision: 2
  scaling: 100
  rounding: 5          ; round to nearest 5 centimes
```

### Accrued interest and high-precision scaling

Interest calculations need much higher precision than display.
Consider daily accrual of a 4-digit interest rate (e.g. 4.567%):

- Annual rate: 0.04567
- Daily rate: 0.04567 / 365 = 0.00012512…
- Applied to 1 penny: 0.00012512… pence per day

To accumulate daily interest without rounding error, a natural
internal scaling is **3,650,000 to the penny** — this allows exact
daily accumulation of any 4-digit annual rate (rate × 10,000 ÷ 365
always yields an integer at this scale).

Higher-precision rates require proportionally larger scaling factors.
The choice of internal precision affects fairness when the accumulated
result is rounded to display precision — different rounding points
produce slightly different AER (Annual Equivalent Rate) values.

### Open questions

- Syntax for declaring rounding rules (nearest 5, banker's rounding,
  etc.).
- Whether internal accrual precision is a commodity property or a
  per-account setting.
- How to express the relationship between a commodity and its minor
  unit variant (e.g. `GBP` vs `GBp`).
- How to declare the conversion/scaling between related commodities
  at different precisions.
