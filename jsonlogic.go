// Package jsonlogic provides a zero-dependency, spec-compliant evaluator for
// JSON Logic rules (see https://jsonlogic.com). Rules are expressed as plain
// JSON and evaluated against arbitrary data, making them safe to store,
// transmit, and generate — including by AI agents.
//
// Basic usage:
//
//	rule := map[string]any{"==": []any{{"var": "name"}, "Alice"}}
//	data := map[string]any{"name": "Alice"}
//	result, err := jsonlogic.Apply(rule, data)
//	// result == true, err == nil
package jsonlogic

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// EvalError is returned when a rule cannot be evaluated. It preserves the
// operator name and the underlying cause so callers can handle errors
// programmatically.
type EvalError struct {
	Operator string
	Cause    error
}

func (e *EvalError) Error() string {
	if e.Operator != "" {
		return fmt.Sprintf("jsonlogic: operator %q: %v", e.Operator, e.Cause)
	}
	return fmt.Sprintf("jsonlogic: %v", e.Cause)
}

func (e *EvalError) Unwrap() error { return e.Cause }

// Apply evaluates a JSON Logic rule against data. Both rule and data must be
// values that are representable as JSON (maps, slices, strings, numbers,
// booleans, nil). Use [ApplyJSON] to evaluate raw JSON bytes.
//
// Returns the computed result, or an [EvalError] if evaluation fails.
func Apply(rule any, data any) (any, error) {
	return apply(rule, data)
}

// ApplyJSON parses rule and data from raw JSON bytes and evaluates the rule.
//
//	result, err := jsonlogic.ApplyJSON([]byte(`{">":[{"var":"age"},17]}`), []byte(`{"age":20}`))
//	// result == true
func ApplyJSON(rule []byte, data []byte) (any, error) {
	var r any
	if err := json.Unmarshal(rule, &r); err != nil {
		return nil, &EvalError{Cause: fmt.Errorf("parsing rule: %w", err)}
	}
	var d any
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, &EvalError{Cause: fmt.Errorf("parsing data: %w", err)}
	}
	return apply(r, d)
}

// MustApply is like Apply but panics on error. Useful in tests and
// initialisation code where the rule is known to be valid.
func MustApply(rule any, data any) any {
	result, err := apply(rule, data)
	if err != nil {
		panic(err)
	}
	return result
}

func apply(rule any, data any) (any, error) {
	// Primitives are returned as-is.
	if rule == nil {
		return nil, nil
	}

	switch v := rule.(type) {
	case bool, float64, string:
		return v, nil
	case []any:
		// Plain arrays: evaluate each element.
		out := make([]any, len(v))
		for i, item := range v {
			res, err := apply(item, data)
			if err != nil {
				return nil, err
			}
			out[i] = res
		}
		return out, nil
	case map[string]any:
		if len(v) == 0 {
			return v, nil
		}
		// A logic rule is a single-key map where the key is the operator.
		if len(v) != 1 {
			// Multi-key maps are returned as plain data (JSON Logic spec §2).
			return v, nil
		}
		for op, rawArgs := range v {
			return evalOperator(op, rawArgs, data)
		}
	}
	return rule, nil
}

func evalOperator(op string, rawArgs any, data any) (any, error) {
	// "var" and "missing" / "missing_some" receive unevaluated args.
	switch op {
	case "var":
		return opVar(rawArgs, data)
	case "missing":
		return opMissing(rawArgs, data)
	case "missing_some":
		return opMissingSome(rawArgs, data)
	}

	// All other operators receive evaluated arguments.
	args, err := evalArgs(rawArgs, data)
	if err != nil {
		return nil, err
	}

	handler, ok := operators[op]
	if !ok {
		return nil, &EvalError{Operator: op, Cause: fmt.Errorf("unknown operator")}
	}
	result, err := handler(args)
	if err != nil {
		return nil, &EvalError{Operator: op, Cause: err}
	}
	return result, nil
}

func evalArgs(rawArgs any, data any) ([]any, error) {
	switch v := rawArgs.(type) {
	case []any:
		out := make([]any, len(v))
		for i, a := range v {
			res, err := apply(a, data)
			if err != nil {
				return nil, err
			}
			out[i] = res
		}
		return out, nil
	default:
		// Single argument — wrap in a slice.
		res, err := apply(rawArgs, data)
		if err != nil {
			return nil, err
		}
		return []any{res}, nil
	}
}

// ---------------------------------------------------------------------------
// var / missing / missing_some
// ---------------------------------------------------------------------------

func opVar(rawArgs any, data any) (any, error) {
	var path string
	var defaultVal any

	switch v := rawArgs.(type) {
	case string:
		path = v
	case []any:
		if len(v) == 0 {
			return data, nil
		}
		path, _ = v[0].(string)
		if len(v) > 1 {
			defaultVal = v[1]
		}
	case float64:
		path = strconv.FormatFloat(v, 'f', -1, 64)
	case nil:
		return data, nil
	}

	if path == "" {
		return data, nil
	}

	val := getPath(data, path)
	if val == nil {
		return defaultVal, nil
	}
	return val, nil
}

func getPath(data any, path string) any {
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		if current == nil {
			return nil
		}
		switch c := current.(type) {
		case map[string]any:
			val, ok := c[part]
			if !ok {
				return nil
			}
			current = val
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil
			}
			current = c[idx]
		default:
			return nil
		}
	}
	return current
}

func opMissing(rawArgs any, data any) (any, error) {
	var keys []any
	switch v := rawArgs.(type) {
	case []any:
		keys = v
	default:
		keys = []any{rawArgs}
	}
	missing := []any{}
	for _, k := range keys {
		path, _ := k.(string)
		if getPath(data, path) == nil {
			missing = append(missing, k)
		}
	}
	return missing, nil
}

func opMissingSome(rawArgs any, data any) (any, error) {
	args, ok := rawArgs.([]any)
	if !ok || len(args) != 2 {
		return nil, fmt.Errorf("missing_some requires [minRequired, keyArray]")
	}
	minRequired := toFloat(args[0])
	keys, _ := args[1].([]any)

	missing := []any{}
	found := 0.0
	for _, k := range keys {
		path, _ := k.(string)
		if getPath(data, path) == nil {
			missing = append(missing, k)
		} else {
			found++
		}
	}
	if found >= minRequired {
		return []any{}, nil
	}
	return missing, nil
}

// ---------------------------------------------------------------------------
// Truthiness helpers (JSON Logic spec)
// ---------------------------------------------------------------------------

// Truthy returns whether a value is considered truthy by JSON Logic rules.
// Empty string, 0, false, nil, and empty slices are falsy; everything else
// is truthy.
func Truthy(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case []any:
		return len(v) != 0
	case map[string]any:
		return true
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		return rv.Len() > 0
	}
	return true
}

func toFloat(val any) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

func toString(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	}
	return fmt.Sprintf("%v", val)
}
