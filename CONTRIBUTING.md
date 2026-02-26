# Contributing

Thank you for your interest in go-jsonlogic!

## Reporting Issues

Please open a GitHub Issue with a minimal reproduction case. If the issue
involves a specific JSON Logic rule, include both the rule and the data as
JSON snippets.

## Pull Requests

1. Fork the repository and create a feature branch.
2. Add or update tests for any changed behaviour.
3. Run `go test ./...` and `go vet ./...` before submitting.
4. Keep pull requests focused — one change per PR.
5. Update `CHANGELOG.md` under an `[Unreleased]` heading.

## Operator Additions

New built-in operators must match the [official JSON Logic specification](https://jsonlogic.com).
Operators that extend beyond the spec belong in a separate package or as
`RegisterOperator` examples in the README rather than in the core library.

## Code Style

- Standard Go formatting (`gofmt`)
- All exported symbols must have GoDoc comments
- Prefer table-driven tests

## License

By contributing you agree that your changes will be licensed under the MIT license.
```

---

## Release and Verification Instructions
```
RELEASE STEPS — GitHub Web UI

1. CREATE THE TAG AND RELEASE
   a. On the repository homepage, click "Releases" in the right sidebar
      (or go to: https://github.com/njchilds90/go-jsonlogic/releases)
   b. Click "Create a new release"
   c. In "Choose a tag", type:   v1.0.0
      Then click "Create new tag: v1.0.0 on publish"
   d. Target branch: main
   e. Release title:   v1.0.0 — Initial Release
   f. In the description box, paste:

## What's included

- Full JSON Logic spec support: equality, comparison, logic, arithmetic,
  string, array, var, missing, missing_some
- `Apply`, `ApplyJSON`, `MustApply`, `Truthy`, `RegisterOperator`
- Structured `*EvalError` type with operator names for programmatic handling
- Zero external dependencies
- Table-driven test suite (50+ cases)
- GitHub Actions CI (Go 1.21 / 1.22 / 1.23)

   g. Leave "Set as the latest release" checked
   h. Click "Publish release"

2. VERIFY ON pkg.go.dev (allow ~10 minutes)
   Visit: https://pkg.go.dev/github.com/njchilds90/go-jsonlogic

   If it hasn't indexed yet, you can force it by running:
   https://pkg.go.dev/github.com/njchilds90/go-jsonlogic@v1.0.0
   (just visiting that URL triggers the indexer)

3. SEMANTIC VERSIONING GUIDANCE
   - Patch releases (v1.0.1): bug fixes, no API changes
   - Minor releases (v1.1.0): new operators or functions, backwards-compatible
   - Major releases (v2.0.0): breaking API changes (avoid for as long as possible)
   - Never retag or delete a published version
