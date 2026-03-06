# Plain Text Accounting Formats

Comparison of plain text accounting formats with [tree-sitter](https://tree-sitter.github.io/)
parsing and syntax highlighting.

## Formats Covered

- **Beancount** — transaction/posting model ([beancount.github.io](https://beancount.github.io/))
- **Goluca** — directed movements with `->` arrows ([tree-sitter-goluca](https://github.com/drummonds/tree-sitter-goluca))
- **PTA** — simplified directed movements ([tree-sitter-pta](https://github.com/drummonds/tree-sitter-pta))
- **Coin** — Go ledger-cli implementation ([mkobetic/coin](https://github.com/mkobetic/coin))

## Documentation

Full format documentation with grammar definitions and examples:
[docs/internal](docs/internal/index.md)

## Install

```bash
go install github.com/drummonds/gts-beancount/cmd/gts-beancount@latest
```

## CLI Usage

```bash
gts-beancount parse ledger.beancount
gts-beancount highlight ledger.beancount
gts-beancount check ledger.beancount
```

## Library Usage

```go
import beancount "github.com/drummonds/gts-beancount"

tree, err := beancount.Parse(src)
ranges, err := beancount.Highlight(src)
golucaSrc, err := beancount.Convert(src)
```

## Build & Test

```bash
task check       # fmt + vet + test
task docs:build  # generate HTML docs
```
