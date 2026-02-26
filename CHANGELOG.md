# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project follows [Semantic Versioning](https://semver.org/).

---

## [1.0.0] - 2026-02-26

### Added

- `Apply(rule, data any)` — evaluate a JSON Logic rule against arbitrary Go values
- `ApplyJSON(rule, data []byte)` — evaluate rules from raw JSON bytes
- `MustApply(rule, data any)` — panic-on-error variant for initialisation code
- `Truthy(val any)` — exported JSON Logic truthiness check
- `RegisterOperator(name, fn)` — extend the evaluator with custom operators
- `*EvalError` — structured error type exposing the failing operator name
- Full spec coverage: equality, comparison, logic, arithmetic, string, array, var, missing, missing_some
- Table-driven test suite with 50+ cases
- GitHub Actions CI matrix (Go 1.21, 1.22, 1.23)
- GoDoc examples for every exported symbol
