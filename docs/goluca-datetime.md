# Goluca Datetime Formats

Goluca datetime handling, covering the current date-only format, the
proposed full datetime extension, and the relationship to ISO 8601 and
RFC 3339.

## Current format

The current tree-sitter grammar supports dates only:

```
YYYY-MM-DD
```

e.g. `2024-01-15`. This is a calendar date with no time component and
no timezone — it identifies the day a transaction occurred.

## Proposed datetime format

The proposed extension adds optional time-of-day and timezone to the
date, following ISO 8601 / RFC 3339:

```abnf
datetime      = date [%s"T" time timezone]

date          = 4DIGIT "-" 2DIGIT "-" 2DIGIT

time          = 2DIGIT ":" 2DIGIT ":" 2DIGIT [fractional]

fractional    = "." 1*6DIGIT       ; up to microsecond precision

timezone      = %s"Z" / tz-offset

tz-offset     = ("+" / "-") 2DIGIT ":" 2DIGIT
```

### Examples

| Datetime | Meaning |
|----------|---------|
| `2024-01-15` | Date only (current format, no ordering within day) |
| `2024-01-15T14:30:00Z` | UTC, second precision |
| `2024-01-15T14:30:00+00:00` | UTC via explicit offset |
| `2024-01-15T14:30:00.123456+01:00` | Microsecond precision, CET |

### Relationship to ISO 8601 and RFC 3339

ISO 8601 defines a broad family of date/time representations. RFC 3339
is a stricter profile of ISO 8601 designed for internet protocols. Goluca
adopts the RFC 3339 subset:

| Constraint | ISO 8601 | RFC 3339 | Goluca |
|------------|----------|----------|--------|
| Date separator | `-` optional | `-` required | `-` required |
| Time separator | `T` or space | `T` (space allowed) | `T` required |
| Timezone | `Z`, `±hh`, `±hhmm`, `±hh:mm` | `Z` or `±hh:mm` | `Z` or `±hh:mm` |
| Fractional seconds | Any precision | Any precision | Up to 6 digits (microsecond) |
| Date-only | Yes | Not specified | Yes (current default) |

Goluca always requires the `T` separator (no space) to keep parsing
unambiguous in a format where payee text follows on the same line.

## Start and end of day

Accounting often needs to distinguish "opening of day" from "close of
day" — for example, an opening balance at the start of 1 January vs a
closing balance at the end of 31 December.

### Proposal: reserved sentinel times

Reserve two specific sub-millisecond times as sentinels:

| Sentinel | Time value | Meaning |
|----------|-----------|---------|
| Start of day | `T00:00:00.001` | Opening instant of the date |
| End of day | `T23:59:59.999` | Closing instant of the date |

All other times within a day are shifted to avoid collision:

- User-supplied times at exactly midnight (`00:00:00.000`) become
  `00:00:00.002` (shifted by 1 ms past the start sentinel).
- User-supplied times at exactly `23:59:59.999` become
  `23:59:59.998` (shifted by 1 ms before the end sentinel).

This carves out two milliseconds from the 86,400,000 available per day
— a negligible loss of precision that gives unambiguous sort order:

```
2024-01-01T00:00:00.001Z   ← start of day (opening balance)
2024-01-01T00:00:00.002Z   ← earliest possible user time
...
2024-01-01T23:59:59.998Z   ← latest possible user time
2024-01-01T23:59:59.999Z   ← end of day (closing balance)
```

### Ordering with date-only entries

When a transaction has no time component (`2024-01-15`), it sorts
*after* start-of-day sentinels and *before* end-of-day sentinels for
that date. Relative ordering among date-only entries on the same day is
undefined (file order).

### Open questions

- Whether sentinels should use UTC or local timezone.
- Whether the sentinel values should be fixed constants or
  configurable per ledger.
- Interaction with knowledge dates: can a knowledge datetime also
  carry start/end-of-day sentinels?
