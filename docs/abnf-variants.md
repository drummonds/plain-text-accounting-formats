# ABNF Standards and Extensions

ABNF (Augmented Backus-Naur Form) is the grammar notation used
throughout this project. The goluca grammars use several features
beyond the base standard. This page documents the relevant standards,
the extensions in use, and why they are needed.

## Standards

### RFC 5234 — base ABNF

The core ABNF specification (2008). Defines the syntax everyone
expects: rule definitions with `=`, repetition with `*`, alternation
with `/`, grouping with `()`, optional with `[]`, and literal strings
in double quotes.

Key properties:

- **Character set**: US-ASCII only (0x00–0x7F).
- **String literals**: Case-insensitive by default —
  `"abc"` matches `ABC`, `Abc`, etc.
- **Numeric literals**: `%d`, `%x`, `%b` for decimal, hex, binary
  code points.
- **Core rules**: `ALPHA`, `DIGIT`, `CRLF`, `WSP`, `DQUOTE`, etc.
  defined in Appendix B.
- **No regex**: Patterns must be expressed as alternations and
  repetitions of individual characters.

> [RFC 5234 — Augmented BNF for Syntax Specifications](https://www.rfc-editor.org/rfc/rfc5234)

### RFC 7405 — case-sensitive strings

An update to RFC 5234 (2014) adding case-sensitive string literals
using the `%s` prefix:

```abnf
; RFC 5234 — case-insensitive (matches "get", "GET", "Get", …)
command = "GET"

; RFC 7405 — case-sensitive (matches only "GET")
command = %s"GET"

; RFC 7405 — explicitly case-insensitive (same as bare quotes)
command = %i"GET"
```

All goluca ABNF uses `%s` for literal strings because accounting
keywords (`commodity`, `open`, `option`) must be exact.

> [RFC 7405 — Case-Sensitive String Support in ABNF](https://www.rfc-editor.org/rfc/rfc7405)

### RFC 5234 + RFC 7405 together

These two RFCs form the "standard ABNF" baseline. The bytestone.uk
hand-written grammar uses only features from these two RFCs.

## The Unicode problem

RFC 5234 is limited to US-ASCII. The goluca arrow operator `→`
(U+2192, RIGHTWARDS ARROW) is a Unicode character and cannot be
expressed as a standard ABNF string literal.

### Options for Unicode in ABNF

**1. Numeric code point (standard)**

```abnf
arrow = %s"->" / %s"//" / %x2192 / %s">"
```

This is valid RFC 5234 — `%x2192` specifies the Unicode code point
directly. It is correct but not human-readable.

**2. UTF-8 byte sequence (standard, ugly)**

```abnf
arrow = %s"->" / %s"//" / %x E2.86.92 / %s">"
```

Encodes the UTF-8 bytes of `→`. Valid but obscure.

**3. Unicode string literal (proposed extension)**

The Leonard–Kyzivat Internet-Draft proposes extending ABNF to allow
Unicode string literals directly:

```abnf
arrow = %s"->" / %s"//" / %s"→" / %s">"
```

This is what the goluca grammars currently use. It reads naturally
but is not yet standardised.

### Goluca's choice

The goluca ABNF uses the proposed Unicode string literal form
(`%s"→"`) because readability matters more than strict RFC compliance
in a documentation grammar. The grammar is a specification for humans
first and a machine input second.

For strict compliance, replace `%s"→"` with `%x2192`. Both describe
the same language.

## tree-sitter extensions

The goluca ABNF is generated from tree-sitter `grammar.js` files by
[tree-sitter2abnf](https://github.com/drummonds/tree-sitter2abnf).
Tree-sitter has concepts with no ABNF equivalent, so the generator
emits extension annotations prefixed with `@`.

### `@pattern` — regular expressions

```abnf
date = @pattern("\\d{4}-\\d{2}-\\d{2}")
```

Standard ABNF would require this to be expanded into character ranges:

```abnf
date = 4DIGIT "-" 2DIGIT "-" 2DIGIT
```

`@pattern` is used when the regex is the canonical form (copied
directly from `grammar.js`) and expanding it would lose clarity or
be impractical (e.g. complex character classes).

### `@token` — indivisible terminals

```abnf
commodity = @token(@prec(1) @pattern("[A-Z][A-Z]+"))
```

Marks a rule as a single token in the tree-sitter parse tree — no
whitespace or child nodes are allowed inside it. In standard ABNF
all terminals are implicitly atomic, so `@token` is informational.

### `@prec` — precedence

```abnf
commodity = @token(@prec(1) @pattern("[A-Z][A-Z]+"))
```

Tree-sitter uses numeric precedence to resolve ambiguities between
overlapping rules (e.g. `commodity` vs `account` both matching
uppercase strings). ABNF has no precedence mechanism — it assumes
an unambiguous grammar. `@prec(n)` documents the tree-sitter
precedence level.

### `@field` — named fields

```abnf
movement = … @field(from) account … @field(to) account …
```

Labels a rule reference with a field name for tree-sitter's API.
The field name has no effect on what the grammar matches — it
annotates the parse tree for programmatic access. No ABNF equivalent.

### `@grammar` and `@extras`

```abnf
; @grammar "goluca"
; @extras (@pattern("\\r"))
```

Metadata comments at the top of the grammar. `@grammar` names the
language; `@extras` lists patterns that are silently skipped between
tokens (like `\r` in line endings). Standard ABNF has no equivalent —
these are tree-sitter configuration.

## Summary of extensions used

| Extension | Source | Standard? | Purpose |
|-----------|--------|-----------|---------|
| `%s"…"` | RFC 7405 | Yes | Case-sensitive string literals |
| `%x2192` | RFC 5234 | Yes | Unicode via numeric code point |
| `%s"→"` | Leonard–Kyzivat draft | No (proposed) | Unicode string literal |
| `@pattern("…")` | tree-sitter2abnf | No | Regex terminal |
| `@token(…)` | tree-sitter2abnf | No | Indivisible terminal |
| `@prec(n)` | tree-sitter2abnf | No | Precedence annotation |
| `@field(name)` | tree-sitter2abnf | No | Named field annotation |
| `@grammar` | tree-sitter2abnf | No | Language name metadata |
| `@extras` | tree-sitter2abnf | No | Skipped-pattern metadata |

## References

- [RFC 5234 — Augmented BNF for Syntax Specifications](https://www.rfc-editor.org/rfc/rfc5234)
- [RFC 7405 — Case-Sensitive String Support in ABNF](https://www.rfc-editor.org/rfc/rfc7405)
- [ABNF and Plain Text Accounting](https://www.bytestone.uk/posts/abnf-and-plain-text-accounting/) — bytestone.uk
- [ABNF syntax comparison](https://www.bytestone.uk/posts/abnf/) — bytestone.uk
- [tree-sitter2abnf](https://github.com/drummonds/tree-sitter2abnf) — grammar.json ↔ ABNF converter
