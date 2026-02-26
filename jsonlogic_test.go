package jsonlogic_test

import (
	"encoding/json"
  "fmt"
	"testing"

	"github.com/njchilds90/go-jsonlogic"
)

func mustJSON(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		panic(err)
	}
	return v
}

func TestApply(t *testing.T) {
	tests := []struct {
		name   string
		rule   string
		data   string
		want   any
		wantOk bool
	}{
		// Primitives
		{"nil rule", `null`, `{}`, nil, true},
		{"bool true", `true`, `{}`, true, true},
		{"bool false", `false`, `{}`, false, true},
		{"number", `42`, `{}`, float64(42), true},
		{"string", `"hello"`, `{}`, "hello", true},

		// var
		{"var simple", `{"var":"name"}`, `{"name":"Alice"}`, "Alice", true},
		{"var nested", `{"var":"a.b"}`, `{"a":{"b":99}}`, float64(99), true},
		{"var missing default", `{"var":["x",7]}`, `{}`, float64(7), true},
		{"var index", `{"var":"0"}`, `["first","second"]`, "first", true},
		{"var empty returns data", `{"var":""}`, `"hello"`, "hello", true},

		// Equality
		{"== equal", `{"==":[1,1]}`, `{}`, true, true},
		{"== soft coerce", `{"==":[1,"1"]}`, `{}`, true, true},
		{"!= not equal", `{"!=":[1,2]}`, `{}`, true, true},
		{"=== strict equal", `{"===":[1,1]}`, `{}`, true, true},
		{"=== strict not equal", `{"===":["1",1]}`, `{}`, false, true},

		// Comparison
		{"> greater", `{">":[2,1]}`, `{}`, true, true},
		{">= gte", `{">=":[2,2]}`, `{}`, true, true},
		{"< less", `{"<":[1,2]}`, `{}`, true, true},
		{"<= lte", `{"<=":[1,2]}`, `{}`, true, true},
		{"< between", `{"<":[1,2,3]}`, `{}`, true, true},
		{"< between fail", `{"<":[1,1,3]}`, `{}`, false, true},

		// Logic
		{"! false", `{"!":[false]}`, `{}`, true, true},
		{"!! truthy", `{"!!":[1]}`, `{}`, true, true},
		{"and true", `{"and":[true,true]}`, `{}`, true, true},
		{"and false", `{"and":[true,false]}`, `{}`, false, true},
		{"or first true", `{"or":[true,false]}`, `{}`, true, true},
		{"or all false", `{"or":[false,false]}`, `{}`, false, true},

		// Arithmetic
		{"+ add", `{"+":[2,3]}`, `{}`, float64(5), true},
		{"+ unary coerce", `{"+":["5"]}`, `{}`, float64(5), true},
		{"- sub", `{"-":[5,2]}`, `{}`, float64(3), true},
		{"- negate", `{"-":[3]}`, `{}`, float64(-3), true},
		{"* mul", `{"*":[2,3]}`, `{}`, float64(6), true},
		{"/ div", `{"/":[6,2]}`, `{}`, float64(3), true},
		{"% mod", `{"%":[7,3]}`, `{}`, float64(1), true},
		{"max", `{"max":[1,3,2]}`, `{}`, float64(3), true},
		{"min", `{"min":[1,3,2]}`, `{}`, float64(1), true},

		// String
		{"cat", `{"cat":["foo","bar"]}`, `{}`, "foobar", true},
		{"in string", `{"in":["foo","foobar"]}`, `{}`, true, true},
		{"in array", `{"in":[2,[1,2,3]]}`, `{}`, true, true},
		{"in array miss", `{"in":[4,[1,2,3]]}`, `{}`, false, true},
		{"substr basic", `{"substr":["hello",1]}`, `{}`, "ello", true},
		{"substr len", `{"substr":["hello",1,3]}`, `{}`, "ell", true},
		{"substr neg", `{"substr":["hello",-2]}`, `{}`, "lo", true},

		// Conditional
		{"if true", `{"if":[true,"yes","no"]}`, `{}`, "yes", true},
		{"if false", `{"if":[false,"yes","no"]}`, `{}`, "no", true},
		{"if no else", `{"if":[false,"yes"]}`, `{}`, nil, true},
		{"?: alias", `{"?:":[true,"a","b"]}`, `{}`, "a", true},

		// Merge
		{"merge", `{"merge":[[1,2],[3,4]]}`, `{}`, []any{float64(1), float64(2), float64(3), float64(4)}, true},

		// missing / missing_some
		{"missing present", `{"missing":["a"]}`, `{"a":1}`, []any{}, true},
		{"missing absent", `{"missing":["b"]}`, `{"a":1}`, []any{"b"}, true},
		{"missing_some enough", `{"missing_some":[1,["a","b"]]}`, `{"a":1}`, []any{}, true},
		{"missing_some not enough", `{"missing_some":[2,["a","b"]]}`, `{"a":1}`, []any{"b"}, true},

		// var in rule
		{"var in comparison", `{">":[{"var":"age"},17]}`, `{"age":20}`, true, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rule := mustJSON(tc.rule)
			data := mustJSON(tc.data)
			got, err := jsonlogic.Apply(rule, data)
			if tc.wantOk && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tc.want) {
				t.Errorf("Apply(%s, %s) = %v (%T), want %v (%T)",
					tc.rule, tc.data, got, got, tc.want, tc.want)
			}
		})
	}
}

func TestApplyJSON(t *testing.T) {
	result, err := jsonlogic.ApplyJSON(
		[]byte(`{">":[{"var":"score"},59]}`),
		[]byte(`{"score":75}`),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestApplyJSONInvalidRule(t *testing.T) {
	_, err := jsonlogic.ApplyJSON([]byte(`{invalid`), []byte(`{}`))
	if err == nil {
		t.Error("expected error for invalid JSON rule")
	}
}

func TestApplyJSONInvalidData(t *testing.T) {
	_, err := jsonlogic.ApplyJSON([]byte(`true`), []byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON data")
	}
}

func TestMustApplyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from division by zero")
		}
	}()
	jsonlogic.MustApply(mustJSON(`{"/":[1,0]}`), nil)
}

func TestDivisionByZero(t *testing.T) {
	_, err := jsonlogic.Apply(mustJSON(`{"/":[1,0]}`), nil)
	if err == nil {
		t.Error("expected error for division by zero")
	}
	e, ok := err.(*jsonlogic.EvalError)
	if !ok {
		t.Fatalf("expected *EvalError, got %T", err)
	}
	if e.Operator != "/" {
		t.Errorf("expected operator '/', got %q", e.Operator)
	}
}

func TestUnknownOperator(t *testing.T) {
	_, err := jsonlogic.Apply(mustJSON(`{"nonexistent":[1]}`), nil)
	if err == nil {
		t.Error("expected error for unknown operator")
	}
}

func TestRegisterOperator(t *testing.T) {
	jsonlogic.RegisterOperator("double", func(args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("need 1 arg")
		}
		f, _ := args[0].(float64)
		return f * 2, nil
	})
	result, err := jsonlogic.Apply(mustJSON(`{"double":[21]}`), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != float64(42) {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestTruthy(t *testing.T) {
	truthy := []any{true, float64(1), "x", []any{1}, map[string]any{"a": 1}}
	falsy := []any{false, float64(0), "", []any{}, nil}

	for _, v := range truthy {
		if !jsonlogic.Truthy(v) {
			t.Errorf("expected %v to be truthy", v)
		}
	}
	for _, v := range falsy {
		if jsonlogic.Truthy(v) {
			t.Errorf("expected %v to be falsy", v)
		}
	}
}

// Example in package documentation
func ExampleApply() {
	rule := map[string]any{
		">": []any{map[string]any{"var": "age"}, 17},
	}
	data := map[string]any{"age": 20}
	result, _ := jsonlogic.Apply(rule, data)
	fmt.Println(result)
	// Output: true
}

func ExampleApplyJSON() {
	result, _ := jsonlogic.ApplyJSON(
		[]byte(`{"==":[{"var":"status"},"active"]}`),
		[]byte(`{"status":"active"}`),
	)
	fmt.Println(result)
	// Output: true
}

func ExampleTruthy() {
	fmt.Println(jsonlogic.Truthy(""))
	fmt.Println(jsonlogic.Truthy("hello"))
	// Output:
	// false
	// true
}

// Needed for test file to compile
import "fmt"
