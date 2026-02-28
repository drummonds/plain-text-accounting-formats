# gts-beancount

Exemplar Go project demonstrating [gotreesitter](https://github.com/odvcencio/gotreesitter) for parsing and syntax-highlighting [beancount](https://beancount.github.io/) files.

Use this as a reference for building parsers for other plain-text formats.

## Install

```bash
go install github.com/drummonds/gts-beancount/cmd/gts-beancount@latest
```

## CLI Usage

```bash
# Print S-expression parse tree
gts-beancount parse ledger.beancount

# Print ANSI-colored output
gts-beancount highlight ledger.beancount

# Check for parse errors
gts-beancount check ledger.beancount
```

## Library Usage

```go
import beancount "github.com/drummonds/gts-beancount"

// Parse source
tree, err := beancount.Parse(src)

// Parse file
bt, err := beancount.ParseFile("ledger.beancount")

// Highlight
ranges, err := beancount.Highlight(src)
```

## Tests

```bash
task check   # fmt + vet + test
task test    # tests only
```

## Adapting for Other Grammars

1. Replace `grammars.BeancountLanguage()` with your target language
2. Write a highlight query matching your grammar's node types
3. Adjust the CLI and tests
