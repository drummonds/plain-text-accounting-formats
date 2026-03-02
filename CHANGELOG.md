# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).
## [0.1.0] - 2026-03-02

 - Starting conversion

## [0.1.0] - 2026-02-28

### Added
- Core parsing API: `Parse`, `ParseFile`, `Language`
- Beancount highlight query covering all directive types
- `Highlight` function returning styled ranges
- `SExpression` helper for tree inspection
- CLI with `parse`, `highlight`, and `check` subcommands
- Test data files covering simple, transaction, and full beancount syntax
