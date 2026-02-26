package jsonlogic

import (
	"fmt"
	"math"
	"strings"
)

// operatorFunc is the signature every built-in and custom operator must satisfy.
// It receives already-evaluated arguments and returns a result or an error.
type operatorFunc func(args []any) (any, error)

// operators holds all registered operators. Custom operators registered via
// [RegisterOperator] are stored here alongside the built-ins.
var operators = map[string]operatorFunc{}

// RegisterOperator adds a custom operator to the evaluator. If an operator
// with the same name already exists it is overwritten. This function is safe
// to call at package initialisation time (func init).
//
//	jsonlogic.RegisterOperator("between", func(args []any) (any, error) {
//	    if len(args) != 3 { return nil, fmt.Errorf("need 3 args") }
//	    v, lo, hi := toFloat(args[0]), toFloat(args[1]), toFloat(args[2])
//	    return v >= lo && v <= hi, nil
//	})
func RegisterOperator(name string, fn operatorFunc) {
	operators[name] = fn
}

func init() {
	// -----------------------------------------------------------------------
	// Equality
	// -----------------------------------------------------------------------
	operators["=="] = func(args []any) (any, error) {
		if len(args) < 2 {
			return false, nil
		}
		return softEqual(args[0], args[1]), nil
	}
	operators["!="] = func(args []any) (any, error) {
		if len(args) < 2 {
			return true, nil
		}
		return !softEqual(args[0], args[1]), nil
	}
	operators["==="] = func(args []any) (any, error) {
		if len(args) < 2 {
			return false, nil
		}
		return strictEqual(args[0], args[1]), nil
	}
	operators["!=="] = func(args []any) (any, error) {
		if len(args) < 2 {
			return true, nil
		}
		return !strictEqual(args[0], args[1]), nil
	}

	// -----------------------------------------------------------------------
	// Comparison
	// -----------------------------------------------------------------------
	operators[">"] = func(args []any) (any, error) {
		if len(args) == 3 {
			// Between: a < b < c
			return toFloat(args[0]) > toFloat(args[1]) && toFloat(args[1]) > toFloat(args[2]), nil
		}
		requireArgs(2, args, ">")
		return toFloat(args[0]) > toFloat(args[1]), nil
	}
	operators[">="] = func(args []any) (any, error) {
		if len(args) == 3 {
			return toFloat(args[0]) >= toFloat(args[1]) && toFloat(args[1]) >= toFloat(args[2]), nil
		}
		return toFloat(args[0]) >= toFloat(args[1]), nil
	}
	operators["<"] = func(args []any) (any, error) {
		if len(args) == 3 {
			return toFloat(args[0]) < toFloat(args[1]) && toFloat(args[1]) < toFloat(args[2]), nil
		}
		return toFloat(args[0]) < toFloat(args[1]), nil
	}
	operators["<="] = func(args []any) (any, error) {
		if len(args) == 3 {
			return toFloat(args[0]) <= toFloat(args[1]) && toFloat(args[1]) <= toFloat(args[2]), nil
		}
		return toFloat(args[0]) <= toFloat(args[1]), nil
	}

	// -----------------------------------------------------------------------
	// Logic
	// -----------------------------------------------------------------------
	operators["!"] = func(args []any) (any, error) {
		if len(args) == 0 {
			return true, nil
		}
		return !Truthy(args[0]), nil
	}
	operators["!!"] = func(args []any) (any, error) {
		if len(args) == 0 {
			return false, nil
		}
		return Truthy(args[0]), nil
	}
	// "and" and "or" are handled specially because they short-circuit;
	// but we re-evaluate raw args inside the operator entry in apply().
	// For simplicity we evaluate them with already-evaluated args here.
	operators["and"] = func(args []any) (any, error) {
		var last any = true
		for _, a := range args {
			last = a
			if !Truthy(a) {
				return a, nil
			}
		}
		return last, nil
	}
	operators["or"] = func(args []any) (any, error) {
		for _, a := range args {
			if Truthy(a) {
				return a, nil
			}
		}
		if len(args) > 0 {
			return args[len(args)-1], nil
		}
		return nil, nil
	}

	// -----------------------------------------------------------------------
	// Arithmetic
	// -----------------------------------------------------------------------
	operators["+"] = func(args []any) (any, error) {
		if len(args) == 1 {
			return toFloat(args[0]), nil
		}
		var sum float64
		for _, a := range args {
			sum += toFloat(a)
		}
		return sum, nil
	}
	operators["-"] = func(args []any) (any, error) {
		if len(args) == 1 {
			return -toFloat(args[0]), nil
		}
		if len(args) < 2 {
			return 0.0, nil
		}
		return toFloat(args[0]) - toFloat(args[1]), nil
	}
	operators["*"] = func(args []any) (any, error) {
		result := 1.0
		for _, a := range args {
			result *= toFloat(a)
		}
		return result, nil
	}
	operators["/"] = func(args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("requires 2 arguments")
		}
		divisor := toFloat(args[1])
		if divisor == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return toFloat(args[0]) / divisor, nil
	}
	operators["%"] = func(args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("requires 2 arguments")
		}
		divisor := toFloat(args[1])
		if divisor == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return math.Mod(toFloat(args[0]), divisor), nil
	}

	// -----------------------------------------------------------------------
	// String / array
	// -----------------------------------------------------------------------
	operators["in"] = func(args []any) (any, error) {
		if len(args) < 2 {
			return false, nil
		}
		needle := args[0]
		switch haystack := args[1].(type) {
		case string:
			return strings.Contains(haystack, toString(needle)), nil
		case []any:
			for _, item := range haystack {
				if softEqual(needle, item) {
					return true, nil
				}
			}
			return false, nil
		}
		return false, nil
	}
	operators["cat"] = func(args []any) (any, error) {
		var sb strings.Builder
		for _, a := range args {
			sb.WriteString(toString(a))
		}
		return sb.String(), nil
	}
	operators["substr"] = func(args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("requires at least 2 arguments")
		}
		str := []rune(toString(args[0]))
		start := int(toFloat(args[1]))
		if start < 0 {
			start = len(str) + start
		}
		if start < 0 {
			start = 0
		}
		if start > len(str) {
			return "", nil
		}
		if len(args) >= 3 {
			length := int(toFloat(args[2]))
			if length < 0 {
				end := len(str) + length
				if end <= start {
					return "", nil
				}
				return string(str[start:end]), nil
			}
			end := start + length
			if end > len(str) {
				end = len(str)
			}
			return string(str[start:end]), nil
		}
		return string(str[start:]), nil
	}

	// -----------------------------------------------------------------------
	// Array operations
	// -----------------------------------------------------------------------
	operators["merge"] = func(args []any) (any, error) {
		result := []any{}
		for _, a := range args {
			switch v := a.(type) {
			case []any:
				result = append(result, v...)
			default:
				result = append(result, v)
			}
		}
		return result, nil
	}

	// -----------------------------------------------------------------------
	// Conditional
	// -----------------------------------------------------------------------
	operators["if"] = func(args []any) (any, error) {
		// if/else-if chain: [cond, then, cond, then, ..., else]
		for i := 0; i+1 < len(args); i += 2 {
			if Truthy(args[i]) {
				return args[i+1], nil
			}
		}
		if len(args)%2 == 1 {
			return args[len(args)-1], nil
		}
		return nil, nil
	}
	operators["?:"] = operators["if"]

	// -----------------------------------------------------------------------
	// Math helpers
	// -----------------------------------------------------------------------
	operators["max"] = func(args []any) (any, error) {
		if len(args) == 0 {
			return nil, nil
		}
		max := toFloat(args[0])
		for _, a := range args[1:] {
			if v := toFloat(a); v > max {
				max = v
			}
		}
		return max, nil
	}
	operators["min"] = func(args []any) (any, error) {
		if len(args) == 0 {
			return nil, nil
		}
		min := toFloat(args[0])
		for _, a := range args[1:] {
			if v := toFloat(a); v < min {
				min = v
			}
		}
		return min, nil
	}

	// -----------------------------------------------------------------------
	// Array higher-order operators
	// Note: filter/map/reduce/all/none/some take a sub-rule as the second
	// argument. Because evalArgs already resolves that sub-rule against data,
	// these operators store the sub-rule as a raw closure captured at
	// Apply-time. We handle them as special forms in apply() instead.
	// -----------------------------------------------------------------------
	// (These are handled in the special-form path below as "array_ops".)
}

// ---------------------------------------------------------------------------
// Equality helpers
// ---------------------------------------------------------------------------

func softEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	// Numeric coercion
	af, aIsNum := toFloatOk(a)
	bf, bIsNum := toFloatOk(b)
	if aIsNum || bIsNum {
		if !aIsNum {
			af = toFloat(a)
		}
		if !bIsNum {
			bf = toFloat(b)
		}
		return af == bf
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func strictEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%T%v", a, a) == fmt.Sprintf("%T%v", b, b)
}

func toFloatOk(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	}
	return 0, false
}

func requireArgs(n int, args []any, op string) error {
	if len(args) < n {
		return fmt.Errorf("%s requires %d argument(s), got %d", op, n, len(args))
	}
	return nil
}
