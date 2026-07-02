// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"testing"
)

func TestArrayOf(t *testing.T) {
	ty := ArrayOf(CoercibleInteger())
	out, err := ty.Call([]any{"1", "2", "3"})
	if err != nil {
		t.Fatalf("array of: %v", err)
	}
	if inspect(out) != "[1, 2, 3]" {
		t.Fatalf("array of got %s", inspect(out))
	}
	// element failure surfaces the element error
	strTy := ArrayOf(StrictInteger())
	if _, err := strTy.Call([]any{1, "x", 3}); err == nil ||
		err.Error() != `"x" violates constraints (type?(Integer, "x") failed)` {
		t.Fatalf("array element fail: %v", err)
	}
	// non-array input
	if _, err := ty.Call(5); err == nil ||
		err.Error() != `5 violates constraints (type?(Array, 5) failed)` {
		t.Fatalf("array non-array: %v", err)
	}
}

func TestHashSchema(t *testing.T) {
	s := NewSchema(
		SchemaKey{Key: "name", Type: StrictString()},
		SchemaKey{Key: "age", Type: CoercibleInteger()},
	)
	out, err := s.Call(map[string]any{"name": "Jane", "age": "30"})
	if err != nil {
		t.Fatalf("schema ok: %v", err)
	}
	m := out.(*Map)
	if v, _ := m.Get(Symbol("name")); v != "Jane" {
		t.Fatalf("schema name: %v", v)
	}
	if v, _ := m.Get(Symbol("age")); v != int64(30) {
		t.Fatalf("schema age: %v", v)
	}

	// missing required key
	if _, err := s.Call(map[string]any{"name": "Jane"}); err == nil ||
		err.Error() != ":age is missing in Hash input" {
		t.Fatalf("schema missing: %v", err)
	}

	// invalid member type
	if _, err := s.Call(map[string]any{"name": 5, "age": "30"}); err == nil ||
		err.Error() != `5 (Integer) has invalid type for :name violates constraints (type?(String, 5) failed)` {
		t.Fatalf("schema bad member: %v", err)
	}

	// extra keys allowed (but dropped) in non-strict schema, matching the gem.
	out2, err := s.Call(map[string]any{"name": "Jane", "age": "30", "extra": 1})
	if err != nil {
		t.Fatalf("schema extra: %v", err)
	}
	if _, ok := out2.(*Map).Get(Symbol("extra")); ok {
		t.Fatal("schema should drop undeclared keys")
	}
	if out2.(*Map).Len() != 2 {
		t.Fatalf("schema output len: %d", out2.(*Map).Len())
	}
}

func TestHashSchemaCoercionMember(t *testing.T) {
	s := NewSchema(SchemaKey{Key: "age", Type: CoercibleInteger()})
	_, err := s.Call(map[string]any{"age": "abc"})
	want := `"abc" (String) has invalid type for :age violates constraints (invalid value for Integer(): "abc" failed)`
	if err == nil || err.Error() != want {
		t.Fatalf("schema coercion member: %v", err)
	}
}

func TestHashSchemaStrictMultiUnknown(t *testing.T) {
	s := NewSchema(SchemaKey{Key: "a", Type: StrictString()}).Strict()
	// *Map input keeps insertion order for the unknown-keys report.
	mm := NewMap()
	mm.Set(Symbol("a"), "x")
	mm.Set(Symbol("c"), 1)
	mm.Set(Symbol("b"), 2)
	_, err := s.Call(mm)
	if err == nil || err.Error() != "unexpected keys [:c, :b] in Hash input" {
		t.Fatalf("multi unknown keys: %v", err)
	}
}

func TestHashSchemaStrict(t *testing.T) {
	s := NewSchema(SchemaKey{Key: "name", Type: StrictString()}).Strict()
	if _, err := s.Call(map[string]any{"name": "Jane", "extra": 1}); err == nil ||
		err.Error() != "unexpected keys [:extra] in Hash input" {
		t.Fatalf("strict schema: %v", err)
	}
	// clean input passes
	if _, err := s.Call(map[string]any{"name": "Jane"}); err != nil {
		t.Fatalf("strict schema clean: %v", err)
	}
}

func TestHashSchemaOptional(t *testing.T) {
	s := NewSchema(
		SchemaKey{Key: "name", Type: StrictString()},
		SchemaKey{Key: "age", Type: CoercibleInteger(), Optional: true},
	)
	// optional key absent -> ok, key omitted
	out, err := s.Call(map[string]any{"name": "Jane"})
	if err != nil {
		t.Fatalf("optional absent: %v", err)
	}
	if _, ok := out.(*Map).Get(Symbol("age")); ok {
		t.Fatal("absent optional key should be omitted")
	}
	// optional key present -> coerced
	out2, _ := s.Call(map[string]any{"name": "Jane", "age": "5"})
	if v, _ := out2.(*Map).Get(Symbol("age")); v != int64(5) {
		t.Fatalf("optional present: %v", v)
	}
}

func TestHashSchemaTypeAndAsType(t *testing.T) {
	s := NewSchema(SchemaKey{Key: "name", Type: StrictString()})
	// AsType composes with combinators.
	opt := s.AsType().Optional()
	if got := call(opt, nil); got != "nil" {
		t.Fatalf("schema optional: %q", got)
	}
	// non-hash input
	if _, err := s.Call(5); err == nil ||
		err.Error() != "5 violates constraints (type?(Hash, 5) failed)" {
		t.Fatalf("schema non-hash: %v", err)
	}
	// s.node() returns the base.
	if s.node() == nil {
		t.Fatal("nil node")
	}
	// ArrayOf a schema.
	arr := ArrayOf(s)
	if _, err := arr.Call([]any{map[string]any{"name": "A"}}); err != nil {
		t.Fatalf("array of schema: %v", err)
	}
}

func TestHashSchemaStringKeyInput(t *testing.T) {
	// Input hash keyed by Symbol, and strict rejecting an unknown symbol key.
	s := NewSchema(SchemaKey{Key: "name", Type: StrictString()}).Strict()
	if _, err := s.Call(map[Symbol]any{"name": "A", "junk": 1}); err == nil {
		t.Fatal("expected unknown symbol key rejection")
	}
	// lookupKey via string key fallback.
	s2 := NewSchema(SchemaKey{Key: "name", Type: StrictString()})
	if _, err := s2.Call(map[Symbol]any{"name": "A"}); err != nil {
		t.Fatalf("symbol-keyed input: %v", err)
	}
	// map[any]any input path
	if _, err := s2.Call(map[any]any{Symbol("name"): "A"}); err != nil {
		t.Fatalf("any-keyed input: %v", err)
	}
	// *Map input with a non-symbol/non-string key present under strict (skipped).
	mm := NewMap()
	mm.Set("name", "A")
	mm.Set([]any{1}, "x") // non-symbol key, ignored by strict scan and passthrough
	if _, err := s.Call(mm); err != nil {
		t.Fatalf("map input non-symbol key: %v", err)
	}
}
