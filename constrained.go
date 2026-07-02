// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math/big"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Constraint is one dry-logic predicate applied by [baseType.Constrained]. The
// map key is the predicate name (gt, gteq, lt, lteq, format, size, min_size,
// max_size, included_in, filled, …) and the value is its argument.
type Constraint struct {
	Name string
	Arg  any
}

// Constrained returns a type that applies each predicate after the base
// coercion/validation (dry-types' `.constrained(...)`). Predicates run in the
// order given; the first failure reports the gem's `<pred>?(<arg>, <val>) failed`
// (or the arity-1 `<pred>?(<val>) failed`) constraint message.
func (b *baseType) Constrained(cs ...Constraint) Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		out, err := base(v)
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			if ok, rule := checkPredicate(c, out); !ok {
				return nil, constraintErr(out, rule)
			}
		}
		return out, nil
	})
}

// checkPredicate evaluates one predicate against v, returning whether it holds
// and the dry-logic rule string to report on failure.
func checkPredicate(c Constraint, v any) (bool, string) {
	switch c.Name {
	case "gt":
		return cmpOK(v, c.Arg, func(d int) bool { return d > 0 }), predRule2("gt", c.Arg, v)
	case "gteq":
		return cmpOK(v, c.Arg, func(d int) bool { return d >= 0 }), predRule2("gteq", c.Arg, v)
	case "lt":
		return cmpOK(v, c.Arg, func(d int) bool { return d < 0 }), predRule2("lt", c.Arg, v)
	case "lteq":
		return cmpOK(v, c.Arg, func(d int) bool { return d <= 0 }), predRule2("lteq", c.Arg, v)
	case "format":
		return formatOK(v, c.Arg), formatRule(c.Arg, v)
	case "size":
		return sizeOK(v, c.Arg, 0), predRule2("size", c.Arg, v)
	case "min_size":
		return sizeOK(v, c.Arg, -1), predRule2("min_size", c.Arg, v)
	case "max_size":
		return sizeOK(v, c.Arg, 1), predRule2("max_size", c.Arg, v)
	case "included_in":
		return includedIn(v, c.Arg), predRule2("included_in", c.Arg, v)
	case "excluded_from":
		return !includedIn(v, c.Arg), predRule2("excluded_from", c.Arg, v)
	case "filled":
		return isFilled(v), "filled?(" + inspect(v) + ") failed"
	case "empty":
		return !isFilled(v), "empty?(" + inspect(v) + ") failed"
	case "eql":
		return valuesEqual(v, c.Arg), predRule2("eql", c.Arg, v)
	}
	return true, ""
}

func predRule2(name string, arg, v any) string {
	return name + "?(" + inspect(arg) + ", " + inspect(v) + ") failed"
}

func formatRule(arg, v any) string {
	return "format?(" + regexpInspect(arg) + ", " + inspect(v) + ") failed"
}

// regexpInspect renders a regexp argument the way Ruby inspects it (/src/).
func regexpInspect(arg any) string {
	switch r := arg.(type) {
	case *regexp.Regexp:
		return "/" + r.String() + "/"
	case string:
		return "/" + r + "/"
	}
	return inspect(arg)
}

// numCompare returns sign(v - arg) for numeric v/arg, and whether both are numeric.
func numCompare(v, arg any) (int, bool) {
	vf, vok := toFloat(v)
	af, aok := toFloat(arg)
	if !vok || !aok {
		return 0, false
	}
	switch {
	case vf < af:
		return -1, true
	case vf > af:
		return 1, true
	default:
		return 0, true
	}
}

func cmpOK(v, arg any, ok func(int) bool) bool {
	d, valid := numCompare(v, arg)
	if !valid {
		return false
	}
	return ok(d)
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case *big.Int:
		f := new(big.Float).SetInt(x)
		r, _ := f.Float64()
		return r, true
	case float64:
		return x, true
	}
	return 0, false
}

// sizeOf returns the Ruby #size / #length of v (string char count or array len).
func sizeOf(v any) (int, bool) {
	switch x := v.(type) {
	case string:
		return utf8.RuneCountInString(x), true
	case []any:
		return len(x), true
	case *Map:
		return x.Len(), true
	}
	return 0, false
}

// sizeOK checks size predicates. mode 0=exact, -1=min, 1=max.
func sizeOK(v, arg any, mode int) bool {
	n, ok := sizeOf(v)
	if !ok {
		return false
	}
	want, wok := toFloat(arg)
	if !wok {
		return false
	}
	switch mode {
	case -1:
		return float64(n) >= want
	case 1:
		return float64(n) <= want
	default:
		return float64(n) == want
	}
}

func formatOK(v, arg any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	switch r := arg.(type) {
	case *regexp.Regexp:
		return r.MatchString(s)
	case string:
		re, err := regexp.Compile(r)
		return err == nil && re.MatchString(s)
	}
	return false
}

func includedIn(v, arg any) bool {
	list, ok := arg.([]any)
	if !ok {
		return false
	}
	for _, e := range list {
		if valuesEqual(v, e) {
			return true
		}
	}
	return false
}

// isFilled reports Ruby truthiness-of-presence: non-nil, non-empty string/array.
func isFilled(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case string:
		return x != ""
	case []any:
		return len(x) != 0
	case *Map:
		return x.Len() != 0
	}
	return true
}

// valuesEqual compares two Ruby values for `==` the way the constraints need
// (numeric cross-type, string, symbol, bool, nil, arrays).
func valuesEqual(a, b any) bool {
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf
		}
		return false
	}
	switch x := a.(type) {
	case string:
		y, ok := b.(string)
		return ok && x == y
	case Symbol:
		y, ok := b.(Symbol)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case nil:
		return b == nil
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !valuesEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// keyList renders a slice of symbol keys as `[:a, :b]` for schema errors, in the
// order given (dry-types preserves the input hash's key order for unknown keys).
func keyList(keys []Symbol) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = ":" + string(k)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
