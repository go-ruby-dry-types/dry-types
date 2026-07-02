// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// paramsTrue / paramsFalse are the exact string sets dry-types' Params bool
// coercion recognizes (Dry::Types::Coercions::Params::TRUE_VALUES / FALSE_VALUES).
var paramsTrue = map[string]bool{
	"1": true, "on": true, "On": true, "ON": true, "t": true,
	"true": true, "True": true, "TRUE": true, "T": true,
	"y": true, "yes": true, "Yes": true, "YES": true, "Y": true,
}

var paramsFalse = map[string]bool{
	"0": true, "off": true, "Off": true, "OFF": true, "f": true,
	"false": true, "False": true, "FALSE": true, "F": true,
	"n": true, "no": true, "No": true, "NO": true, "N": true,
}

// kernelInteger reproduces Ruby's Kernel#Integer used by Coercible::Integer:
// it trims surrounding whitespace, honors 0x/0o/0b prefixes, allows underscore
// digit separators, and coerces numeric inputs. On failure it returns the gem's
// message. (Coercible feeds strings and numerics here.)
func kernelInteger(v any) (any, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case *big.Int:
		return x, nil
	case float64:
		return int64(math.Trunc(x)), nil
	case bool:
		return nil, coercionErr("can't convert " + boolClass(x) + " into Integer")
	case nil:
		return nil, coercionErr("can't convert nil into Integer")
	case string:
		return parseKernelInteger(x)
	}
	return nil, coercionErr("can't convert " + valueClass(v) + " into Integer")
}

// parseKernelInteger mirrors Integer(str): base prefixes, underscores, sign,
// full-string match, whitespace trim.
func parseKernelInteger(s string) (any, error) {
	t := strings.TrimSpace(s)
	fail := func() (any, error) {
		return nil, coercionErr("invalid value for Integer(): " + rubyStringInspect(s))
	}
	if t == "" {
		return fail()
	}
	neg := false
	body := t
	if body[0] == '+' || body[0] == '-' {
		neg = body[0] == '-'
		body = body[1:]
	}
	base := 10
	switch {
	case strings.HasPrefix(body, "0x"), strings.HasPrefix(body, "0X"):
		base, body = 16, body[2:]
	case strings.HasPrefix(body, "0o"), strings.HasPrefix(body, "0O"):
		base, body = 8, body[2:]
	case strings.HasPrefix(body, "0b"), strings.HasPrefix(body, "0B"):
		base, body = 2, body[2:]
	}
	// Underscores are allowed only between digits; reject leading/trailing/double.
	if body == "" || strings.HasPrefix(body, "_") || strings.HasSuffix(body, "_") ||
		strings.Contains(body, "__") {
		return fail()
	}
	clean := strings.ReplaceAll(body, "_", "")
	n, ok := new(big.Int).SetString(clean, base)
	if !ok {
		return fail()
	}
	if neg {
		n.Neg(n)
	}
	if n.IsInt64() {
		return n.Int64(), nil
	}
	return n, nil
}

// kernelFloat reproduces Ruby's Kernel#Float used by Coercible::Float: strict —
// the whole string must be a valid float; whitespace is trimmed; numerics pass.
func kernelFloat(v any) (any, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case *big.Int:
		f := new(big.Float).SetInt(x)
		r, _ := f.Float64()
		return r, nil
	case nil:
		return nil, coercionErr("can't convert nil into Float")
	case bool:
		return nil, coercionErr("can't convert " + boolClass(x) + " into Float")
	case string:
		t := strings.TrimSpace(x)
		clean := strings.ReplaceAll(t, "_", "")
		f, err := strconv.ParseFloat(clean, 64)
		if err != nil || t == "" {
			return nil, coercionErr("invalid value for Float(): " + rubyStringInspect(x))
		}
		return f, nil
	}
	return nil, coercionErr("can't convert " + valueClass(v) + " into Float")
}

// kernelString reproduces Coercible::String: `input.to_s`. nil→"", true→"true",
// numerics/symbols/arrays via their Ruby to_s.
func kernelString(v any) (any, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case nil:
		return "", nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	case Symbol:
		return string(x), nil
	case int:
		return strconv.Itoa(x), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case *big.Int:
		return x.String(), nil
	case float64:
		return formatFloat(x), nil
	case []any:
		return arrayToS(x), nil
	}
	return inspect(v), nil
}

// arrayToS renders []any the way Ruby's Array#to_s (== inspect) does.
func arrayToS(a []any) string {
	parts := make([]string, len(a))
	for i, e := range a {
		parts[i] = inspect(e)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// kernelSymbol reproduces Coercible::Symbol: `input.to_sym` — only strings and
// symbols respond; anything else raises the NoMethodError-shaped message.
func kernelSymbol(v any) (any, error) {
	switch x := v.(type) {
	case Symbol:
		return x, nil
	case string:
		return Symbol(x), nil
	}
	return nil, coercionErr("undefined method 'to_sym' for an instance of " + valueClass(v))
}

// paramsInteger: Params::Integer. Empty string → nil (dry-types maps "" to nil
// before coercing, but Params::Integer itself raises on ""); the caller wires the
// empty→nil rule where the gem does. Here it defers to Integer() semantics with a
// non-base-prefixed decimal-only string per params (form values are decimal).
func paramsInteger(v any) (any, error) {
	s, ok := v.(string)
	if !ok {
		return kernelInteger(v)
	}
	t := strings.TrimSpace(s)
	if t == "" {
		return nil, coercionErr("invalid value for Integer(): " + rubyStringInspect(s))
	}
	// Params uses Integer(input, 10): base-10 only, no 0x prefixes.
	neg := false
	body := t
	if body != "" && (body[0] == '+' || body[0] == '-') {
		neg = body[0] == '-'
		body = body[1:]
	}
	if body == "" || strings.HasPrefix(body, "_") || strings.HasSuffix(body, "_") ||
		strings.Contains(body, "__") {
		return nil, coercionErr("invalid value for Integer(): " + rubyStringInspect(s))
	}
	clean := strings.ReplaceAll(body, "_", "")
	n, okp := new(big.Int).SetString(clean, 10)
	if !okp {
		return nil, coercionErr("invalid value for Integer(): " + rubyStringInspect(s))
	}
	if neg {
		n.Neg(n)
	}
	if n.IsInt64() {
		return n.Int64(), nil
	}
	return n, nil
}

// paramsFloat: Params::Float — Float() on the string.
func paramsFloat(v any) (any, error) { return kernelFloat(v) }

// paramsBool: Params::Bool. Recognizes the TRUE_VALUES / FALSE_VALUES sets;
// passes through real bools; raises the gem's "cannot be coerced" message.
func paramsBool(v any) (any, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case string:
		if paramsTrue[x] {
			return true, nil
		}
		if paramsFalse[x] {
			return false, nil
		}
		// The gem's message uses the negated literal it was checking against.
		return nil, coercionErr(x + " cannot be coerced to false")
	}
	return nil, coercionErr(inspect(v) + " cannot be coerced to false")
}

// paramsNil: Params::Nil — "" → nil; a real nil passes; anything else raises.
func paramsNil(v any) (any, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case string:
		if x == "" {
			return nil, nil
		}
	}
	return nil, coercionErr(inspect(v) + " is not nil")
}

// paramsSymbol: Params::Symbol — to_sym.
func paramsSymbol(v any) (any, error) { return kernelSymbol(v) }

// coerceDate parses a date the way dry-types' Date coercion (Date.parse) does for
// ISO forms, passing through an existing Date/Time.
func coerceDate(v any) (any, error) {
	switch x := v.(type) {
	case Date:
		return x, nil
	case time.Time:
		return Date{Year: x.Year(), Month: int(x.Month()), Day: x.Day()}, nil
	case string:
		if d, ok := parseDate(x); ok {
			return d, nil
		}
		return nil, coercionErr("invalid date")
	}
	return nil, coercionErr("invalid date")
}

// coerceTime parses a time the way dry-types' Time coercion (Time.parse) does.
func coerceTime(v any) (any, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	case Date:
		return x.toTime(), nil
	case string:
		if tm, ok := parseTime(x); ok {
			return tm, nil
		}
		return nil, coercionErr("no time information in " + rubyStringInspect(x))
	}
	return nil, coercionErr("no time information in " + inspect(v))
}

// coerceDateTime parses a DateTime (same underlying parse as Time here).
func coerceDateTime(v any) (any, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	case Date:
		return x.toTime(), nil
	case string:
		if tm, ok := parseTime(x); ok {
			return tm, nil
		}
		return nil, coercionErr("invalid date")
	}
	return nil, coercionErr("invalid date")
}

func parseDate(s string) (Date, bool) {
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339, "2006-01-02T15:04:05"} {
		if tm, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
			return Date{Year: tm.Year(), Month: int(tm.Month()), Day: tm.Day()}, true
		}
	}
	return Date{}, false
}

func parseTime(s string) (time.Time, bool) {
	for _, layout := range []string{
		time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05",
		"2006-01-02 15:04", "2006-01-02", "15:04:05", "15:04",
	} {
		if tm, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
			return tm, true
		}
	}
	return time.Time{}, false
}

// boolClass renders the Ruby class name of a bool for coercion error messages.
func boolClass(b bool) string {
	if b {
		return "TrueClass"
	}
	return "FalseClass"
}

// valueClass renders the Ruby class name of a value for coercion error messages.
func valueClass(v any) string {
	switch v.(type) {
	case nil:
		return "NilClass"
	case bool:
		return boolClass(v.(bool))
	case string:
		return "String"
	case Symbol:
		return "Symbol"
	case int, int32, int64, *big.Int:
		return "Integer"
	case float64:
		return "Float"
	case []any:
		return "Array"
	case *Map, map[string]any, map[Symbol]any, map[any]any:
		return "Hash"
	case Date:
		return "Date"
	case time.Time:
		return "Time"
	}
	return "Object"
}
