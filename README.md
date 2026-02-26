# go-jsonlogic

[![CI](https://github.com/njchilds90/go-jsonlogic/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-jsonlogic/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/go-jsonlogic.svg)](https://pkg.go.dev/github.com/njchilds90/go-jsonlogic)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A zero-dependency, spec-compliant Go implementation of [JSON Logic](https://jsonlogic.com) — the portable, language-agnostic standard for expressing business rules as plain JSON.

Write rules once. Evaluate them anywhere: frontend (JavaScript), backend (Go), or by an AI agent that generates rules dynamically.

---

## Why JSON Logic?

- Rules are **plain JSON** — safe to store in a database, send over a network, or generate programmatically
- **No code execution risk** — perfect for user-defined rules or agent-generated logic
- **Portable** — the same rule works in JavaScript, Python, Ruby, PHP, and now Go
- **Deterministic** — identical input always produces identical output

---

## Installation
```bash
go get github.com/njchilds90/go-jsonlogic
```

---

## Quick Start
```go
package main

import (
    "fmt"
    "github.com/njchilds90/go-jsonlogic"
)

func main() {
    // Is the user old enough?
    rule := map[string]any{
        ">": []any{map[string]any{"var": "age"}, 17},
    }
    data := map[string]any{"age": 20}

    result, err := jsonlogic.Apply(rule, data)
    fmt.Println(result, err) // true <nil>
}
```

### From raw JSON (most common in production)
```go
result, err := jsonlogic.ApplyJSON(
    []byte(`{"and":[{">":[{"var":"age"},17]},{"==":[{"var":"country"},"US"]}]}`),
    []byte(`{"age":20,"country":"US"}`),
)
// result == true
```

---

## Supported Operators

| Category     | Operators                                      |
|--------------|------------------------------------------------|
| Equality     | `==` `!=` `===` `!==`                          |
| Comparison   | `>` `>=` `<` `<=` (with 3-arg between form)   |
| Logic        | `!` `!!` `and` `or` `if` / `?:`               |
| Arithmetic   | `+` `-` `*` `/` `%` `max` `min`               |
| String/Array | `cat` `substr` `in` `merge`                   |
| Data access  | `var` `missing` `missing_some`                |

All operators match the [official JSON Logic test suite](https://jsonlogic.com/tests.json).

---

## Custom Operators

Extend the evaluator with your own operators at startup:
```go
func init() {
    jsonlogic.RegisterOperator("between", func(args []any) (any, error) {
        if len(args) != 3 {
            return nil, fmt.Errorf("between requires 3 arguments")
        }
        v := toFloat(args[0])
        lo := toFloat(args[1])
        hi := toFloat(args[2])
        return v >= lo && v <= hi, nil
    })
}

// {"between":[5, 1, 10]}  →  true
```

---

## Error Handling

All errors are typed as `*jsonlogic.EvalError`, which exposes the operator name for programmatic handling:
```go
result, err := jsonlogic.Apply(rule, data)
if err != nil {
    var evalErr *jsonlogic.EvalError
    if errors.As(err, &evalErr) {
        log.Printf("failed at operator %q: %v", evalErr.Operator, evalErr.Cause)
    }
}
```

---

## Use Cases

- **Feature flags** — store rollout rules in your database; evaluate in Go with zero overhead
- **Access control / policy evaluation** — portable rules that match frontend behaviour exactly
- **Discount and pricing engines** — business team writes rules; engineers deploy without code changes
- **Form validation** — share validation rules between your API and client
- **AI agent rule generation** — agents produce JSON Logic rules that are safe to evaluate without `eval()`

---

## Design Principles

- Zero external runtime dependencies
- Deterministic: identical inputs always produce identical outputs
- All errors are typed and structured — no string matching required
- Extensible: add custom operators without forking
- Pure functions; no hidden global state beyond the operator registry
- Idiomatic Go naming and patterns throughout

---

## License

MIT — see [LICENSE](LICENSE)
