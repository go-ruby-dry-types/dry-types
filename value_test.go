// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math"
	"math/big"
	"testing"
	"time"
)

func TestValidTry(t *testing.T) {
	if !Valid(StrictInteger(), 5) {
		t.Fatal("valid should be true")
	}
	if Valid(StrictInteger(), "x") {
		t.Fatal("valid should be false")
	}
	r := Try(StrictInteger(), 5)
	if !r.Success() || r.Failure() || r.Input() != 5 || r.Error() != nil {
		t.Fatalf("try success: %+v", r)
	}
	r2 := Try(StrictInteger(), "x")
	if r2.Success() || !r2.Failure() || r2.Error() == nil {
		t.Fatalf("try failure: %+v", r2)
	}
	if r2.Input() != "x" {
		t.Fatalf("try failure input: %v", r2.Input())
	}
}

func TestMapBasics(t *testing.T) {
	m := NewMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 3) // dedup replace
	if m.Len() != 2 {
		t.Fatalf("len: %d", m.Len())
	}
	if v, ok := m.Get("a"); !ok || v != 3 {
		t.Fatalf("get a: %v", v)
	}
	if _, ok := m.Get("missing"); ok {
		t.Fatal("missing should be absent")
	}
	if len(m.Pairs()) != 2 {
		t.Fatal("pairs len")
	}
	// non-comparable key appends (never dedups).
	m.Set([]any{1}, "x")
	m.Set([]any{1}, "y")
	if m.Len() != 4 {
		t.Fatalf("noncomparable len: %d", m.Len())
	}
	if _, ok := m.Get([]any{1}); ok {
		t.Fatal("noncomparable get should miss")
	}
	// zero-value Map (nil index) Set path.
	var z Map
	z.Set("k", 1)
	if z.Len() != 1 {
		t.Fatal("zero map set")
	}
}

func TestAsMap(t *testing.T) {
	if _, ok := asMap(NewMap()); !ok {
		t.Fatal("*Map")
	}
	if m, ok := asMap(map[string]any{"b": 2, "a": 1}); !ok || m.Len() != 2 {
		t.Fatal("map[string]any")
	}
	if m, ok := asMap(map[Symbol]any{"b": 2, "a": 1}); !ok || m.Len() != 2 {
		t.Fatal("map[Symbol]any")
	}
	if m, ok := asMap(map[any]any{"a": 1}); !ok || m.Len() != 1 {
		t.Fatal("map[any]any")
	}
	if _, ok := asMap(5); ok {
		t.Fatal("non-map")
	}
}

func TestAsBigInt(t *testing.T) {
	for _, v := range []any{int(1), int32(1), int64(1), big.NewInt(1)} {
		if _, ok := asBigInt(v); !ok {
			t.Fatalf("asBigInt %T", v)
		}
	}
	if _, ok := asBigInt("x"); ok {
		t.Fatal("asBigInt string")
	}
}

func TestInspect(t *testing.T) {
	bi, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	cases := []struct {
		v    any
		want string
	}{
		{nil, "nil"},
		{true, "true"},
		{false, "false"},
		{"hi", `"hi"`},
		{"a\nb\tc\r\"\\", `"a\nb\tc\r\"\\"`},
		{Symbol("s"), ":s"},
		{int(5), "5"},
		{int32(5), "5"},
		{int64(5), "5"},
		{bi, "123456789012345678901234567890"},
		{3.5, "3.5"},
		{2.0, "2.0"},
		{math.Inf(1), "+Inf"},
		{[]any{1, "a"}, `[1, "a"]`},
		{Date{2026, 1, 2}, "#<Date: 2026-01-02>"},
	}
	for _, c := range cases {
		if got := inspect(c.v); got != c.want {
			t.Fatalf("inspect(%v) = %q want %q", c.v, got, c.want)
		}
	}
	// *Map inspect: symbol-key shorthand and string-key => form.
	m := NewMap()
	m.Set(Symbol("name"), "Jane")
	m.Set("k", 1)
	if got := inspect(m); got != `{name: "Jane", "k" => 1}` {
		t.Fatalf("map inspect: %q", got)
	}
	// time.Time and unknown fallthrough.
	_ = inspect(time.Now())
	_ = inspect(struct{}{})
}

func TestFormatFloat(t *testing.T) {
	if formatFloat(2.0) != "2.0" {
		t.Fatal("2.0")
	}
	if formatFloat(3.5) != "3.5" {
		t.Fatal("3.5")
	}
}

func TestDateHelpers(t *testing.T) {
	d := Date{2026, 7, 2}
	if d.String() != "2026-07-02" {
		t.Fatalf("date string: %s", d.String())
	}
	tm := d.toTime()
	if tm.Year() != 2026 || tm.Month() != 7 || tm.Day() != 2 {
		t.Fatalf("toTime: %v", tm)
	}
}

func TestValueClass(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{nil, "NilClass"},
		{true, "TrueClass"},
		{false, "FalseClass"},
		{"x", "String"},
		{Symbol("s"), "Symbol"},
		{5, "Integer"},
		{int64(5), "Integer"},
		{big.NewInt(5), "Integer"},
		{3.5, "Float"},
		{[]any{}, "Array"},
		{NewMap(), "Hash"},
		{map[string]any{}, "Hash"},
		{Date{}, "Date"},
		{time.Now(), "Time"},
		{struct{}{}, "Object"},
	}
	for _, c := range cases {
		if got := valueClass(c.v); got != c.want {
			t.Fatalf("valueClass(%v)=%s want %s", c.v, got, c.want)
		}
	}
}

func TestKernelStringArray(t *testing.T) {
	if got := arrayToS([]any{1, "a", true}); got != `[1, "a", true]` {
		t.Fatalf("arrayToS: %s", got)
	}
}

func TestRegexpInspect(t *testing.T) {
	if got := regexpInspect("^a"); got != "/^a/" {
		t.Fatalf("regexpInspect string: %s", got)
	}
}

func TestToFloat(t *testing.T) {
	cases := []struct {
		v  any
		ok bool
		f  float64
	}{
		{int(5), true, 5},
		{int32(5), true, 5},
		{int64(5), true, 5},
		{big.NewInt(5), true, 5},
		{3.5, true, 3.5},
		{"x", false, 0},
	}
	for _, c := range cases {
		f, ok := toFloat(c.v)
		if ok != c.ok || (ok && f != c.f) {
			t.Fatalf("toFloat(%v)=%v,%v want %v,%v", c.v, f, ok, c.f, c.ok)
		}
	}
}

func TestConstrainedIntTypes(t *testing.T) {
	// Drive gt with int32/int64/big.Int operands to cover toFloat branches in a
	// constraint context (the number is the coerced value).
	ty := NominalInteger().Constrained(Constraint{"gt", 1})
	for _, v := range []any{int32(5), int64(5), big.NewInt(5)} {
		if _, err := ty.Call(v); err != nil {
			t.Fatalf("gt on %T: %v", v, err)
		}
	}
	// eql comparing a float value against a non-numeric fails cleanly.
	if valuesEqual(3.5, "x") {
		t.Fatal("float vs string should be unequal")
	}
}

func TestValuesEqual(t *testing.T) {
	cases := []struct {
		a, b any
		want bool
	}{
		{5, 5.0, true},
		{5, "5", false},
		{"a", "a", true},
		{"a", "b", false},
		{"a", 5, false},
		{Symbol("s"), Symbol("s"), true},
		{Symbol("s"), Symbol("t"), false},
		{Symbol("s"), "s", false},
		{true, true, true},
		{true, false, false},
		{true, "x", false},
		{nil, nil, true},
		{nil, 5, false},
		{[]any{1, 2}, []any{1, 2}, true},
		{[]any{1, 2}, []any{1, 3}, false},
		{[]any{1}, []any{1, 2}, false},
		{[]any{1}, 5, false},
		{Date{2026, 1, 1}, Date{2026, 1, 1}, false}, // unhandled type -> default false
	}
	for _, c := range cases {
		if got := valuesEqual(c.a, c.b); got != c.want {
			t.Fatalf("valuesEqual(%v,%v)=%v want %v", c.a, c.b, got, c.want)
		}
	}
}
