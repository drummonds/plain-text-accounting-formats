# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).## [0.1.1] - 2026-03-03

 - Colouring syntax comparison this is what this is meant to show## [0.1.3] - 2026-03-03

 - Trying to update docs git hub pages## [0.1.5] - 2026-03-05

 - Improved documentation and push to github pages## [0.1.7] - 2026-03-09

 - Separate documentation site## [0.1.9] - 2026-03-14

 - Updating docs## [0.1.11] - 2026-03-15

 - lint rules fix
## [0.1.12] - 2026-03-15

 - extending go-luca

## [0.1.10] - 2026-03-14

 - updating docs

## [0.1.8] - 2026-03-14

 - Just rebasing

## [0.1.6] - 2026-03-07

 - Release prep

## [0.1.4] - 2026-03-05

 - Changing documentation

## [0.1.2] - 2026-03-03

 - Working syntax

## [0.2.0] - 2026-03-03

### Added
- Goluca parsing and syntax highlighting via `GolucaParse`, `GolucaHighlight`
- Side-by-side beancount/goluca display in docs page
- Forked gotreesitter with goluca and pta grammar support

### Changed
- Switched dependency from `odvcencio/gotreesitter` to `drummonds/gotreesitter` v0.5.3

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
