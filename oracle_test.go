// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// rubyBin locates a usable `ruby` with the dry-types gem once. The oracle tests
// skip themselves when either is absent (the qemu cross-arch lanes, the Windows
// lane, and any host without the gem), so the deterministic suite alone drives
// the 100% coverage gate there.
func rubyBin(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping dry-types oracle")
	}
	// Gate on RUBY_VERSION >= "4.0" and gem availability, matching the mandate.
	probe := exec.Command(path, "-e",
		`exit((RUBY_VERSION.split(".").first.to_i >= 4) ? 0 : 3)`)
	if err := probe.Run(); err != nil {
		t.Skip("ruby < 4.0; skipping dry-types oracle")
	}
	check := exec.Command(path, "-e", `require "dry-types"`)
	if err := check.Run(); err != nil {
		t.Skip("dry-types gem not installed; skipping oracle")
	}
	return path
}

// rubyOutcome runs a dry-types expression against the gem and returns a canonical
// "OK <inspect>" / "ERR <message>" string for comparison with the Go outcome.
func rubyOutcome(t *testing.T, bin, expr string) string {
	t.Helper()
	script := `
$stdout.binmode
require "dry-types"
module Types; include Dry.Types(); end
begin
  r = (` + expr + `)
  print "OK "
  print r.inspect
rescue => e
  print "ERR "
  print e.message
end
`
	out, err := exec.Command(bin, "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error for %q: %v\n%s", expr, err, out)
	}
	return strings.TrimSpace(string(out))
}

// goOutcome canonicalizes a Go Type application the same way rubyOutcome does.
func goOutcome(ty Type, input any) string {
	out, err := ty.Call(input)
	if err != nil {
		return "ERR " + err.Error()
	}
	return "OK " + inspect(out)
}

// TestOracleScalars checks success values AND error messages of the scalar
// coercion/validation types match the gem byte-for-byte.
func TestOracleScalars(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		expr string // Ruby expression
		ty   Type   // equivalent Go type
		in   any    // Go input
	}{
		{`Types::Strict::Integer[5]`, StrictInteger(), 5},
		{`Types::Strict::Integer["5"]`, StrictInteger(), "5"},
		{`Types::Strict::String["x"]`, StrictString(), "x"},
		{`Types::Strict::String[5]`, StrictString(), 5},
		{`Types::Strict::Float[3.5]`, StrictFloat(), 3.5},
		{`Types::Strict::Float[3]`, StrictFloat(), 3},
		{`Types::Strict::Bool[true]`, StrictBool(), true},
		{`Types::Strict::Bool["true"]`, StrictBool(), "true"},
		{`Types::Strict::Symbol[:s]`, StrictSymbol(), Symbol("s")},
		{`Types::Strict::Nil[nil]`, StrictNil(), nil},
		{`Types::Coercible::Integer["5"]`, CoercibleInteger(), "5"},
		{`Types::Coercible::Integer["0xff"]`, CoercibleInteger(), "0xff"},
		{`Types::Coercible::Integer["1_000"]`, CoercibleInteger(), "1_000"},
		{`Types::Coercible::Integer["abc"]`, CoercibleInteger(), "abc"},
		{`Types::Coercible::Integer[nil]`, CoercibleInteger(), nil},
		{`Types::Coercible::Integer[3.9]`, CoercibleInteger(), 3.9},
		{`Types::Coercible::Float["3.5"]`, CoercibleFloat(), "3.5"},
		{`Types::Coercible::Float["3.5x"]`, CoercibleFloat(), "3.5x"},
		{`Types::Coercible::Float[3]`, CoercibleFloat(), 3},
		{`Types::Coercible::String[5]`, CoercibleString(), 5},
		{`Types::Coercible::String[nil]`, CoercibleString(), nil},
		{`Types::Coercible::String[true]`, CoercibleString(), true},
		{`Types::Coercible::String[[1, 2]]`, CoercibleString(), []any{1, 2}},
		{`Types::Coercible::Symbol["foo"]`, CoercibleSymbol(), "foo"},
		{`Types::Coercible::Symbol[5]`, CoercibleSymbol(), 5},
		{`Types::Params::Integer["5"]`, ParamsInteger(), "5"},
		{`Types::Params::Integer[""]`, ParamsInteger(), ""},
		{`Types::Params::Integer["3.5"]`, ParamsInteger(), "3.5"},
		{`Types::Params::Float["3"]`, ParamsFloat(), "3"},
		{`Types::Params::Float[""]`, ParamsFloat(), ""},
		{`Types::Params::Bool["1"]`, ParamsBool(), "1"},
		{`Types::Params::Bool["true"]`, ParamsBool(), "true"},
		{`Types::Params::Bool["yes"]`, ParamsBool(), "yes"},
		{`Types::Params::Bool["0"]`, ParamsBool(), "0"},
		{`Types::Params::Bool["off"]`, ParamsBool(), "off"},
		{`Types::Params::Bool["xyz"]`, ParamsBool(), "xyz"},
		{`Types::Params::Nil[""]`, ParamsNil(), ""},
		{`Types::Params::Nil["x"]`, ParamsNil(), "x"},
		{`Types::Params::Symbol["foo"]`, ParamsSymbol(), "foo"},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			want := rubyOutcome(t, bin, c.expr)
			got := goOutcome(c.ty, c.in)
			if got != want {
				t.Fatalf("expr %s:\n  go   = %q\n  ruby = %q", c.expr, got, want)
			}
		})
	}
}

// TestOracleConstraints checks constrained-type outcomes match the gem.
func TestOracleConstraints(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		expr string
		ty   Type
		in   any
	}{
		{`Types::Strict::Integer.constrained(gt: 18)[20]`, StrictInteger().Constrained(Constraint{"gt", 18}), 20},
		{`Types::Strict::Integer.constrained(gt: 18)[10]`, StrictInteger().Constrained(Constraint{"gt", 18}), 10},
		{`Types::Strict::Integer.constrained(gteq: 18)[17]`, StrictInteger().Constrained(Constraint{"gteq", 18}), 17},
		{`Types::Strict::Integer.constrained(lt: 5)[10]`, StrictInteger().Constrained(Constraint{"lt", 5}), 10},
		{`Types::Strict::Integer.constrained(lteq: 5)[6]`, StrictInteger().Constrained(Constraint{"lteq", 5}), 6},
		{`Types::Strict::String.constrained(format: /^a/)["xyz"]`, StrictString().Constrained(Constraint{"format", regexp.MustCompile("^a")}), "xyz"},
		{`Types::Strict::String.constrained(size: 3)["ab"]`, StrictString().Constrained(Constraint{"size", 3}), "ab"},
		{`Types::Strict::String.constrained(min_size: 3)["ab"]`, StrictString().Constrained(Constraint{"min_size", 3}), "ab"},
		{`Types::Strict::String.constrained(max_size: 3)["abcd"]`, StrictString().Constrained(Constraint{"max_size", 3}), "abcd"},
		{`Types::Strict::String.constrained(included_in: %w[a b c])["z"]`, StrictString().Constrained(Constraint{"included_in", []any{"a", "b", "c"}}), "z"},
		{`Types::Strict::String.constrained(filled: true)[""]`, StrictString().Constrained(Constraint{"filled", true}), ""},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			want := rubyOutcome(t, bin, c.expr)
			got := goOutcome(c.ty, c.in)
			if got != want {
				t.Fatalf("expr %s:\n  go   = %q\n  ruby = %q", c.expr, got, want)
			}
		})
	}
}

// TestOracleCombinators checks optional/default/enum/sum/array/schema.
func TestOracleCombinators(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		expr string
		ty   Type
		in   any
	}{
		{`Types::Strict::Integer.optional[nil]`, StrictInteger().Optional(), nil},
		{`Types::Strict::Integer.optional["x"]`, StrictInteger().Optional(), "x"},
		{`Types::Strict::Integer.default(0)[Dry::Types::Undefined]`, StrictInteger().Default(0), Undefined},
		{`Types::String.enum("draft", "published")["x"]`, StrictString().Enum("draft", "published"), "x"},
		{`(Types::Strict::Integer | Types::Strict::String)[5]`, StrictInteger().Or(StrictString()), 5},
		{`(Types::Strict::Integer | Types::Strict::String)[1.5]`, StrictInteger().Or(StrictString()), 1.5},
		{`Types::Array.of(Types::Coercible::Integer)[["1", "2", "3"]]`, ArrayOf(CoercibleInteger()), []any{"1", "2", "3"}},
		{`Types::Array.of(Types::Strict::Integer)[[1, "x"]]`, ArrayOf(StrictInteger()), []any{1, "x"}},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			want := rubyOutcome(t, bin, c.expr)
			got := goOutcome(c.ty, c.in)
			if got != want {
				t.Fatalf("expr %s:\n  go   = %q\n  ruby = %q", c.expr, got, want)
			}
		})
	}
}

// TestOracleSchemaErrors checks the schema error messages (missing / invalid /
// unknown keys) against the gem.
func TestOracleSchemaErrors(t *testing.T) {
	bin := rubyBin(t)
	schema := NewSchema(
		SchemaKey{Key: "name", Type: StrictString()},
		SchemaKey{Key: "age", Type: CoercibleInteger()},
	)
	rubyDef := `s = Types::Hash.schema(name: Types::Strict::String, age: Types::Coercible::Integer)`

	cases := []struct {
		expr string
		in   any
		ty   Type
	}{
		{rubyDef + `; s[{name: "Jane", age: "30"}]`, map[string]any{"name": "Jane", "age": "30"}, schema},
		{rubyDef + `; s[{name: "Jane"}]`, map[string]any{"name": "Jane"}, schema},
		{rubyDef + `; s[{name: 5, age: "30"}]`, map[string]any{"name": 5, "age": "30"}, schema},
		{rubyDef + `; s[{name: "Jane", age: "abc"}]`, map[string]any{"name": "Jane", "age": "abc"}, schema},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			want := rubyOutcome(t, bin, c.expr)
			got := goOutcome(c.ty, c.in)
			if got != want {
				t.Fatalf("expr %s:\n  go   = %q\n  ruby = %q", c.expr, got, want)
			}
		})
	}

	// strict schema unknown keys
	strict := schema.Strict()
	want := rubyOutcome(t, bin, rubyDef+`.strict; s[{name: "Jane", extra: 1}]`)
	got := goOutcome(strict, map[string]any{"name": "Jane", "extra": 1})
	if got != want {
		t.Fatalf("strict schema:\n  go   = %q\n  ruby = %q", got, want)
	}
}
