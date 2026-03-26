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
go install codeberg.org/hum3/plain-text-accounting-formats/cmd/ptaf@latest
```

## CLI Usage

```bash
gts-beancount parse ledger.beancount
gts-beancount highlight ledger.beancount
gts-beancount check ledger.beancount
```

## Library Usage

```go
import ptaf "codeberg.org/hum3/plain-text-accounting-formats"

tree, err := beancount.Parse(src)
ranges, err := beancount.Highlight(src)
golucaSrc, err := beancount.Convert(src)
```

## Build & Test

```bash
task check       # fmt + vet + test
task docs:build  # generate HTML docs
```

## How documentation fits together

Go library and CLI for parsing and syntax-highlighting plain-text accounting formats (beancount, goluca, PTA) using [gotreesitter](https://github.com/drummonds/gotreesitter).

So this is linked to other sites:

- go-luca (using the go luca format to convert text journals to ledgers DB/in memory)
- This web site showing how to use gotreesitter for syntax highlighting, conversion from one format to another and import and export examples, event streaming?
- (https://github.com/drummonds/pta2svg) Producing graphs
- https://github.com/drummonds/gotreesitter has embedded the beancount, goluca and pta formats
    - https://github.com/drummonds/tree-sitter-goluca
    - https://github.com/drummonds/tree-sitter-pta
- tree-sitter2abnf
- bytestone discussion of formats

## Links

| | |
|---|---|
| Documentation | https://h3-pta-formats.statichost.page/ |
| Source (Codeberg) | https://codeberg.org/hum3/plain-text-accounting-formats |
| Mirror (GitHub) | https://github.com/drummonds/plain-text-accounting-formats |
