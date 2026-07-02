// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math/big"
	"reflect"
	"regexp"
	"testing"
	"time"
)

// call is a test helper: applies t to input and returns a printable outcome —
// the coerced value's Ruby inspect on success, or the error message on failure.
func call(t Type, input any) string {
	out, err := t.Call(input)
	if err != nil {
		return err.Error()
	}
	if out == nil {
		return "nil"
	}
	return inspect(out)
}

func TestStrict(t *testing.T) {
	cases := []struct {
		name string
		ty   Type
		in   any
		want string
	}{
		{"int-ok", StrictInteger(), 5, "5"},
		{"int-bad", StrictInteger(), "5", `"5" violates constraints (type?(Integer, "5") failed)`},
		{"int64-ok", StrictInteger(), int64(5), "5"},
		{"int32-ok", StrictInteger(), int32(5), "5"},
		{"bigint-ok", StrictInteger(), big.NewInt(5), "5"},
		{"float-ok", StrictFloat(), 3.5, "3.5"},
		{"float-bad", StrictFloat(), 3, `3 violates constraints (type?(Float, 3) failed)`},
		{"string-ok", StrictString(), "x", `"x"`},
		{"string-bad", StrictString(), 5, `5 violates constraints (type?(String, 5) failed)`},
		{"symbol-ok", StrictSymbol(), Symbol("s"), ":s"},
		{"symbol-bad", StrictSymbol(), "s", `"s" violates constraints (type?(Symbol, "s") failed)`},
		{"bool-true", StrictBool(), true, "true"},
		{"bool-false", StrictBool(), false, "false"},
		{"bool-bad", StrictBool(), "true", `"true" violates constraints (type?(FalseClass, "true") failed)`},
		{"nil-ok", StrictNil(), nil, "nil"},
		{"nil-bad", StrictNil(), 5, `5 violates constraints (type?(NilClass, 5) failed)`},
		{"array-ok", StrictArray(), []any{1}, "[1]"},
		{"array-bad", StrictArray(), 5, `5 violates constraints (type?(Array, 5) failed)`},
		{"hash-ok", StrictHash(), NewMap(), ""},
		{"hash-bad", StrictHash(), 5, `5 violates constraints (type?(Hash, 5) failed)`},
		{"date-ok", StrictDate(), Date{2026, 7, 2}, "#<Date: 2026-07-02>"},
		{"date-bad", StrictDate(), 5, `5 violates constraints (type?(Date, 5) failed)`},
		{"time-ok", StrictTime(), time.Now(), ""},
		{"time-bad", StrictTime(), 5, `5 violates constraints (type?(Time, 5) failed)`},
		{"datetime-bad", StrictDateTime(), 5, `5 violates constraints (type?(DateTime, 5) failed)`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := call(c.ty, c.in)
			if c.want != "" && got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestNominal(t *testing.T) {
	// Nominal types apply no constraint: any input passes through unchanged.
	cases := []struct {
		ty Type
		in any
	}{
		{NominalInteger(), "not-an-int"},
		{NominalString(), 5},
		{NominalFloat(), "x"},
		{NominalSymbol(), 1},
		{NominalBool(), "x"},
		{NominalArray(), 5},
		{NominalHash(), 5},
	}
	for _, c := range cases {
		out, err := c.ty.Call(c.in)
		if err != nil {
			t.Fatalf("nominal raised: %v", err)
		}
		if !reflect.DeepEqual(out, c.in) {
			t.Fatalf("nominal changed value: %v -> %v", c.in, out)
		}
	}
}

func TestCoercible(t *testing.T) {
	cases := []struct {
		name string
		ty   Type
		in   any
		want string
	}{
		{"int-str", CoercibleInteger(), "5", "5"},
		{"int-ws", CoercibleInteger(), " 5 ", "5"},
		{"int-under", CoercibleInteger(), "1_000", "1000"},
		{"int-hex", CoercibleInteger(), "0xff", "255"},
		{"int-oct", CoercibleInteger(), "0o17", "15"},
		{"int-bin", CoercibleInteger(), "0b101", "5"},
		{"int-neg", CoercibleInteger(), "-7", "-7"},
		{"int-float", CoercibleInteger(), 3.9, "3"},
		{"int-int", CoercibleInteger(), 5, "5"},
		{"int-int64", CoercibleInteger(), int64(5), "5"},
		{"int-int32", CoercibleInteger(), int32(5), "5"},
		{"int-big", CoercibleInteger(), big.NewInt(9), "9"},
		{"int-bignum", CoercibleInteger(), "123456789012345678901234567890", "123456789012345678901234567890"},
		{"int-bad", CoercibleInteger(), "abc", `invalid value for Integer(): "abc"`},
		{"int-empty", CoercibleInteger(), "", `invalid value for Integer(): ""`},
		{"int-nil", CoercibleInteger(), nil, "can't convert nil into Integer"},
		{"int-true", CoercibleInteger(), true, "can't convert TrueClass into Integer"},
		{"int-false", CoercibleInteger(), false, "can't convert FalseClass into Integer"},
		{"int-badunder", CoercibleInteger(), "1__0", `invalid value for Integer(): "1__0"`},
		{"int-trailunder", CoercibleInteger(), "10_", `invalid value for Integer(): "10_"`},
		{"int-leadunder", CoercibleInteger(), "_10", `invalid value for Integer(): "_10"`},
		{"int-array", CoercibleInteger(), []any{1}, "can't convert Array into Integer"},
		{"float-str", CoercibleFloat(), "3.5", "3.5"},
		{"float-int", CoercibleFloat(), 3, "3.0"},
		{"float-int64", CoercibleFloat(), int64(3), "3.0"},
		{"float-int32", CoercibleFloat(), int32(3), "3.0"},
		{"float-big", CoercibleFloat(), big.NewInt(3), "3.0"},
		{"float-float", CoercibleFloat(), 2.5, "2.5"},
		{"float-junk", CoercibleFloat(), "3.5x", `invalid value for Float(): "3.5x"`},
		{"float-empty", CoercibleFloat(), "", `invalid value for Float(): ""`},
		{"float-nil", CoercibleFloat(), nil, "can't convert nil into Float"},
		{"float-true", CoercibleFloat(), true, "can't convert TrueClass into Float"},
		{"float-array", CoercibleFloat(), []any{1}, "can't convert Array into Float"},
		{"str-int", CoercibleString(), 5, `"5"`},
		{"str-int64", CoercibleString(), int64(5), `"5"`},
		{"str-int32", CoercibleString(), int32(5), `"5"`},
		{"str-big", CoercibleString(), big.NewInt(5), `"5"`},
		{"str-float", CoercibleString(), 3.5, `"3.5"`},
		{"str-nil", CoercibleString(), nil, `""`},
		{"str-true", CoercibleString(), true, `"true"`},
		{"str-false", CoercibleString(), false, `"false"`},
		{"str-sym", CoercibleString(), Symbol("foo"), `"foo"`},
		{"str-str", CoercibleString(), "x", `"x"`},
		{"str-array", CoercibleString(), []any{1, 2}, `"[1, 2]"`},
		{"str-other", CoercibleString(), Date{2026, 1, 1}, `"#<Date: 2026-01-01>"`},
		{"sym-str", CoercibleSymbol(), "foo", ":foo"},
		{"sym-sym", CoercibleSymbol(), Symbol("foo"), ":foo"},
		{"sym-int", CoercibleSymbol(), 5, "undefined method 'to_sym' for an instance of Integer"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := call(c.ty, c.in); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestParams(t *testing.T) {
	cases := []struct {
		name string
		ty   Type
		in   any
		want string
	}{
		{"int", ParamsInteger(), "5", "5"},
		{"int-neg", ParamsInteger(), "-5", "-5"},
		{"int-plus", ParamsInteger(), "+5", "5"},
		{"int-under", ParamsInteger(), "1_000", "1000"},
		{"int-big", ParamsInteger(), "123456789012345678901234567890", "123456789012345678901234567890"},
		{"int-empty", ParamsInteger(), "", `invalid value for Integer(): ""`},
		{"int-float", ParamsInteger(), "3.5", `invalid value for Integer(): "3.5"`},
		{"int-badunder", ParamsInteger(), "1__0", `invalid value for Integer(): "1__0"`},
		{"int-trailunder", ParamsInteger(), "10_", `invalid value for Integer(): "10_"`},
		{"int-onlysign", ParamsInteger(), "-", `invalid value for Integer(): "-"`},
		{"int-numeric", ParamsInteger(), 5, "5"},
		{"float", ParamsFloat(), "3.5", "3.5"},
		{"float-int", ParamsFloat(), "3", "3.0"},
		{"float-empty", ParamsFloat(), "", `invalid value for Float(): ""`},
		{"bool-1", ParamsBool(), "1", "true"},
		{"bool-true", ParamsBool(), "true", "true"},
		{"bool-yes", ParamsBool(), "yes", "true"},
		{"bool-on", ParamsBool(), "on", "true"},
		{"bool-y", ParamsBool(), "y", "true"},
		{"bool-t", ParamsBool(), "t", "true"},
		{"bool-TRUE", ParamsBool(), "TRUE", "true"},
		{"bool-0", ParamsBool(), "0", "false"},
		{"bool-false", ParamsBool(), "false", "false"},
		{"bool-no", ParamsBool(), "no", "false"},
		{"bool-realtrue", ParamsBool(), true, "true"},
		{"bool-realfalse", ParamsBool(), false, "false"},
		{"bool-bad", ParamsBool(), "xyz", "xyz cannot be coerced to false"},
		{"bool-badtype", ParamsBool(), 5, "5 cannot be coerced to false"},
		{"nil-empty", ParamsNil(), "", "nil"},
		{"nil-nil", ParamsNil(), nil, "nil"},
		{"nil-nonempty", ParamsNil(), "x", `"x" is not nil`},
		{"nil-nonstr", ParamsNil(), 5, "5 is not nil"},
		{"sym", ParamsSymbol(), "foo", ":foo"},
		{"sym-bad", ParamsSymbol(), 5, "undefined method 'to_sym' for an instance of Integer"},
		{"date", ParamsDate(), "2026-07-02", "#<Date: 2026-07-02>"},
		{"date-slash", ParamsDate(), "2026/07/02", "#<Date: 2026-07-02>"},
		{"date-fromdate", ParamsDate(), Date{2026, 7, 2}, "#<Date: 2026-07-02>"},
		{"date-bad", ParamsDate(), "notadate", "invalid date"},
		{"date-badtype", ParamsDate(), 5, "invalid date"},
		{"time-bad", ParamsTime(), "notatime", `no time information in "notatime"`},
		{"time-badtype", ParamsTime(), 5, "no time information in 5"},
		{"datetime-bad", ParamsDateTime(), "notadate", "invalid date"},
		{"datetime-badtype", ParamsDateTime(), 5, "invalid date"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := call(c.ty, c.in); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestParamsTemporalOK(t *testing.T) {
	// Date/Time/DateTime successful coercions from time-bearing inputs.
	if _, err := ParamsTime().Call("2026-07-02 10:30:00"); err != nil {
		t.Fatalf("params time: %v", err)
	}
	if _, err := ParamsTime().Call(Date{2026, 7, 2}); err != nil {
		t.Fatalf("params time from date: %v", err)
	}
	if _, err := ParamsTime().Call(time.Now()); err != nil {
		t.Fatalf("params time from time: %v", err)
	}
	if _, err := ParamsDateTime().Call("2026-07-02T10:30:00"); err != nil {
		t.Fatalf("params datetime: %v", err)
	}
	if _, err := ParamsDateTime().Call(time.Now()); err != nil {
		t.Fatalf("params datetime from time: %v", err)
	}
	if _, err := ParamsDateTime().Call(Date{2026, 7, 2}); err != nil {
		t.Fatalf("params datetime from date: %v", err)
	}
	if _, err := ParamsDate().Call(time.Now()); err != nil {
		t.Fatalf("params date from time: %v", err)
	}
	if _, err := ParamsDate().Call("2026-07-02T10:00:00"); err != nil {
		t.Fatalf("params date from datetime str: %v", err)
	}
	// Date parsed from RFC3339 string.
	if _, err := ParamsDate().Call("2026-07-02T10:00:00Z"); err != nil {
		t.Fatalf("params date rfc3339: %v", err)
	}
}

func TestJSON(t *testing.T) {
	if got := call(JSONDate(), "2026-07-02"); got != "#<Date: 2026-07-02>" {
		t.Fatalf("json date: %q", got)
	}
	if got := call(JSONSymbol(), "foo"); got != ":foo" {
		t.Fatalf("json symbol: %q", got)
	}
	if got := call(JSONNil(), nil); got != "nil" {
		t.Fatalf("json nil: %q", got)
	}
	if _, err := JSONTime().Call("2026-07-02 10:00:00"); err != nil {
		t.Fatalf("json time: %v", err)
	}
	if _, err := JSONDateTime().Call("2026-07-02T10:00:00"); err != nil {
		t.Fatalf("json datetime: %v", err)
	}
}

func TestOptional(t *testing.T) {
	ty := StrictInteger().Optional()
	if got := call(ty, nil); got != "nil" {
		t.Fatalf("optional nil: %q", got)
	}
	if got := call(ty, 5); got != "5" {
		t.Fatalf("optional val: %q", got)
	}
	if got := call(ty, "x"); got != `"x" violates constraints (type?(Integer, "x") failed)` {
		t.Fatalf("optional bad: %q", got)
	}
}

func TestDefault(t *testing.T) {
	ty := StrictInteger().Default(0)
	if got := call(ty, Undefined); got != "0" {
		t.Fatalf("default: %q", got)
	}
	if got := call(ty, 5); got != "5" {
		t.Fatalf("default val: %q", got)
	}
	fnTy := StrictInteger().DefaultFn(func() any { return 42 })
	if got := call(fnTy, Undefined); got != "42" {
		t.Fatalf("default fn: %q", got)
	}
	if got := call(fnTy, 7); got != "7" {
		t.Fatalf("default fn val: %q", got)
	}
}

func TestConstructor(t *testing.T) {
	ty := StrictString().Constructor(func(v any) any {
		s, _ := v.(string)
		// trim
		for len(s) > 0 && s[0] == ' ' {
			s = s[1:]
		}
		for len(s) > 0 && s[len(s)-1] == ' ' {
			s = s[:len(s)-1]
		}
		return s
	})
	if got := call(ty, "  hi  "); got != `"hi"` {
		t.Fatalf("constructor: %q", got)
	}
}

func TestMeta(t *testing.T) {
	base := StrictInteger().(*baseType)
	if len(base.GetMeta()) != 0 {
		t.Fatal("expected empty meta")
	}
	m := base.Meta(map[string]any{"foo": "bar"})
	if m.GetMeta()["foo"] != "bar" {
		t.Fatalf("meta: %v", m.GetMeta())
	}
	// Meta merges.
	m2 := m.Meta(map[string]any{"baz": 1})
	if m2.GetMeta()["foo"] != "bar" || m2.GetMeta()["baz"] != 1 {
		t.Fatalf("meta merge: %v", m2.GetMeta())
	}
	// Meta carries through a combinator.
	opt := m.Optional()
	if opt.GetMeta()["foo"] != "bar" {
		t.Fatalf("meta not carried: %v", opt.GetMeta())
	}
}

func TestOr(t *testing.T) {
	ty := StrictInteger().Or(StrictString())
	if got := call(ty, 5); got != "5" {
		t.Fatalf("sum int: %q", got)
	}
	if got := call(ty, "x"); got != `"x"` {
		t.Fatalf("sum str: %q", got)
	}
	if got := call(ty, 1.5); got != `1.5 violates constraints (type?(String, 1.5) failed)` {
		t.Fatalf("sum bad: %q", got)
	}
}

func TestEnum(t *testing.T) {
	ty := StrictString().Enum("draft", "published")
	if got := call(ty, "draft"); got != `"draft"` {
		t.Fatalf("enum ok: %q", got)
	}
	if got := call(ty, "x"); got != `"x" violates constraints (included_in?(["draft", "published"], "x") failed)` {
		t.Fatalf("enum bad: %q", got)
	}
	// Enum on a failing base surfaces the base error.
	ityp := StrictInteger().Enum(1, 2)
	if got := call(ityp, "x"); got != `"x" violates constraints (type?(Integer, "x") failed)` {
		t.Fatalf("enum base fail: %q", got)
	}
}

func TestConstrained(t *testing.T) {
	re := regexp.MustCompile("^a")
	cases := []struct {
		name string
		ty   Type
		in   any
		want string
	}{
		{"gt-ok", StrictInteger().Constrained(Constraint{"gt", 18}), 20, "20"},
		{"gt-bad", StrictInteger().Constrained(Constraint{"gt", 18}), 10, "10 violates constraints (gt?(18, 10) failed)"},
		{"gteq-ok", StrictInteger().Constrained(Constraint{"gteq", 18}), 18, "18"},
		{"gteq-bad", StrictInteger().Constrained(Constraint{"gteq", 18}), 17, "17 violates constraints (gteq?(18, 17) failed)"},
		{"lt-bad", StrictInteger().Constrained(Constraint{"lt", 5}), 10, "10 violates constraints (lt?(5, 10) failed)"},
		{"lt-ok", StrictInteger().Constrained(Constraint{"lt", 5}), 3, "3"},
		{"lteq-bad", StrictInteger().Constrained(Constraint{"lteq", 5}), 6, "6 violates constraints (lteq?(5, 6) failed)"},
		{"lteq-ok", StrictInteger().Constrained(Constraint{"lteq", 5}), 5, "5"},
		{"format-ok", StrictString().Constrained(Constraint{"format", re}), "abc", `"abc"`},
		{"format-bad", StrictString().Constrained(Constraint{"format", re}), "xyz", `"xyz" violates constraints (format?(/^a/, "xyz") failed)`},
		{"format-str-ok", StrictString().Constrained(Constraint{"format", "^a"}), "abc", `"abc"`},
		{"format-nonstr", StrictInteger().Or(StrictString()).Constrained(Constraint{"format", re}), 5, `5 violates constraints (format?(/^a/, 5) failed)`},
		{"size-bad", StrictString().Constrained(Constraint{"size", 3}), "ab", `"ab" violates constraints (size?(3, "ab") failed)`},
		{"size-ok", StrictString().Constrained(Constraint{"size", 3}), "abc", `"abc"`},
		{"minsize-bad", StrictString().Constrained(Constraint{"min_size", 3}), "ab", `"ab" violates constraints (min_size?(3, "ab") failed)`},
		{"minsize-ok", StrictString().Constrained(Constraint{"min_size", 3}), "abcd", `"abcd"`},
		{"maxsize-bad", StrictString().Constrained(Constraint{"max_size", 3}), "abcd", `"abcd" violates constraints (max_size?(3, "abcd") failed)`},
		{"maxsize-ok", StrictString().Constrained(Constraint{"max_size", 3}), "ab", `"ab"`},
		{"size-arr", StrictArray().Constrained(Constraint{"size", 2}), []any{1, 2}, "[1, 2]"},
		{"included-ok", StrictString().Constrained(Constraint{"included_in", []any{"a", "b", "c"}}), "a", `"a"`},
		{"included-bad", StrictString().Constrained(Constraint{"included_in", []any{"a", "b", "c"}}), "z", `"z" violates constraints (included_in?(["a", "b", "c"], "z") failed)`},
		{"excluded-bad", StrictString().Constrained(Constraint{"excluded_from", []any{"a"}}), "a", `"a" violates constraints (excluded_from?(["a"], "a") failed)`},
		{"excluded-ok", StrictString().Constrained(Constraint{"excluded_from", []any{"a"}}), "b", `"b"`},
		{"filled-bad", StrictString().Constrained(Constraint{"filled", true}), "", `"" violates constraints (filled?("") failed)`},
		{"filled-ok", StrictString().Constrained(Constraint{"filled", true}), "x", `"x"`},
		{"empty-bad", StrictString().Constrained(Constraint{"empty", true}), "x", `"x" violates constraints (empty?("x") failed)`},
		{"empty-ok", StrictString().Constrained(Constraint{"empty", true}), "", `""`},
		{"eql-ok", StrictInteger().Constrained(Constraint{"eql", 5}), 5, "5"},
		{"eql-bad", StrictInteger().Constrained(Constraint{"eql", 5}), 6, "6 violates constraints (eql?(5, 6) failed)"},
		{"base-fail", StrictInteger().Constrained(Constraint{"gt", 1}), "x", `"x" violates constraints (type?(Integer, "x") failed)`},
		{"unknown-pred", StrictInteger().Constrained(Constraint{"bogus", 1}), 5, "5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := call(c.ty, c.in); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestConstrainedEdge(t *testing.T) {
	// Non-numeric value in a numeric predicate fails cleanly.
	ty := NominalInteger().Constrained(Constraint{"gt", 1})
	if got := call(ty, "x"); got != `"x" violates constraints (gt?(1, "x") failed)` {
		t.Fatalf("gt nonnum: %q", got)
	}
	// Non-numeric predicate arg.
	ty2 := StrictInteger().Constrained(Constraint{"gt", "x"})
	if got := call(ty2, 5); got != `5 violates constraints (gt?("x", 5) failed)` {
		t.Fatalf("gt nonnum arg: %q", got)
	}
	// size on non-sized value fails.
	ty3 := NominalInteger().Constrained(Constraint{"size", 1})
	if got := call(ty3, 5); got != "5 violates constraints (size?(1, 5) failed)" {
		t.Fatalf("size nonsize: %q", got)
	}
	// size with non-numeric arg fails.
	ty3b := StrictString().Constrained(Constraint{"size", "x"})
	if _, err := ty3b.Call("abc"); err == nil {
		t.Fatal("expected size bad-arg failure")
	}
	// included_in with non-array arg fails.
	ty4 := StrictString().Constrained(Constraint{"included_in", "x"})
	if _, err := ty4.Call("a"); err == nil {
		t.Fatal("expected included_in bad-arg failure")
	}
	// format with bad regexp arg type fails.
	ty5 := StrictString().Constrained(Constraint{"format", 5})
	if _, err := ty5.Call("a"); err == nil {
		t.Fatal("expected format bad-arg failure")
	}
	// format with invalid regexp source string fails.
	ty5b := StrictString().Constrained(Constraint{"format", "("})
	if _, err := ty5b.Call("a"); err == nil {
		t.Fatal("expected format invalid-regexp failure")
	}
	// filled on array / map / nil.
	fa := NominalArray().Constrained(Constraint{"filled", true})
	if _, err := fa.Call([]any{}); err == nil {
		t.Fatal("expected empty array filled failure")
	}
	if _, err := fa.Call([]any{1}); err != nil {
		t.Fatalf("nonempty array should pass filled: %v", err)
	}
	fm := NominalHash().Constrained(Constraint{"filled", true})
	if _, err := fm.Call(NewMap()); err == nil {
		t.Fatal("expected empty map filled failure")
	}
	fn := NominalInteger().Constrained(Constraint{"filled", true})
	if _, err := fn.Call(nil); err == nil {
		t.Fatal("expected nil filled failure")
	}
	if _, err := fn.Call(5); err != nil {
		t.Fatalf("int should pass filled: %v", err)
	}
	// empty on map.
	em := NominalHash().Constrained(Constraint{"empty", true})
	if _, err := em.Call(NewMap()); err != nil {
		t.Fatalf("empty map should pass empty: %v", err)
	}
	// regexpInspect fallback branch (non-regexp, non-string arg rendered).
	_ = formatRule(5, "a")
	// size on a Map (sized).
	sm := NominalHash().Constrained(Constraint{"size", 0})
	if _, err := sm.Call(NewMap()); err != nil {
		t.Fatalf("empty map size 0: %v", err)
	}
}
