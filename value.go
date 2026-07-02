// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package drytypes is a pure-Go (CGO-free) MRI-faithful reimplementation of the
// Ruby dry-types gem: a composable type system with coercion, constraints, and
// combinators. Every type is a [Type] — a value you build from a constructor
// (Strict, Coercible, Params, JSON, Nominal, …) and refine with combinator
// methods (Optional, Default, Constrained, Enum, Or, Constructor, …). Applying a
// type to an input coerces and validates it, returning the coerced value or an
// error whose message is byte-identical to the dry-types gem's.
//
// # Ruby value model
//
// Ruby values are represented by the same small, fixed set of Go types the
// go-ruby-* ecosystem uses, so a host (go-embedded-ruby / rbgo) maps its object
// graph to and from this package with no glue:
//
//	Ruby            Go
//	----            --
//	nil             nil
//	true / false    bool
//	Integer         int64, *big.Int (int/int32 accepted on input)
//	Float           float64
//	String          string
//	Symbol          Symbol
//	Array           []any
//	Hash            *Map (ordered), map[string]any / map[Symbol]any (input)
//	Date            Date
//	Time / DateTime Time
//
// A host callable (a default block or a constructor function) is a Go closure of
// type [func(any) any] or [Callable].
package drytypes

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
)

// Symbol is a Ruby Symbol (`:name`).
type Symbol string

// Date is a Ruby Date (a calendar day with no time-of-day). It is the coercion
// target of the *::Date types. Stored as year/month/day plus the parsed
// underlying time for formatting.
type Date struct {
	Year  int
	Month int
	Day   int
}

// String renders the Date the way Ruby's Date#to_s / inspect renders it
// (ISO-8601 `YYYY-MM-DD`).
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// toTime is the midnight-UTC time.Time view of the Date, used internally.
func (d Date) toTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// Time is a Ruby Time / DateTime. It wraps a Go time.Time; the *::DateTime types
// coerce into the same shape (dry-types coerces both via Date._parse-style
// parsing and yields a Time-like object the host distinguishes by its own class
// mapping).
type Time = time.Time

// Pair is one entry of an ordered mapping.
type Pair struct {
	Key any
	Val any
}

// Map is an insertion-ordered Ruby Hash. Schema coercion yields a *Map so key
// order round-trips.
type Map struct {
	pairs []Pair
	index map[any]int
}

// NewMap returns an empty ordered Map.
func NewMap() *Map { return &Map{index: map[any]int{}} }

// Len reports the number of entries.
func (m *Map) Len() int { return len(m.pairs) }

// Pairs returns the entries in insertion order. The slice must not be mutated.
func (m *Map) Pairs() []Pair { return m.pairs }

// Set inserts or replaces the entry for key (comparable keys deduplicate).
func (m *Map) Set(key, val any) {
	if m.index == nil {
		m.index = map[any]int{}
	}
	if comparableKey(key) {
		if i, ok := m.index[key]; ok {
			m.pairs[i].Val = val
			return
		}
		m.index[key] = len(m.pairs)
	}
	m.pairs = append(m.pairs, Pair{Key: key, Val: val})
}

// Get returns the value for a comparable key and whether it was present.
func (m *Map) Get(key any) (any, bool) {
	if comparableKey(key) {
		if i, ok := m.index[key]; ok {
			return m.pairs[i].Val, true
		}
	}
	return nil, false
}

func comparableKey(key any) bool {
	switch key.(type) {
	case []any, *Map:
		return false
	}
	return true
}

// asMap normalizes any hash-shaped input into an ordered *Map. It accepts *Map,
// map[string]any, map[Symbol]any and map[any]any; it returns (nil, false) for
// anything else so callers can raise the proper type error.
func asMap(v any) (*Map, bool) {
	switch h := v.(type) {
	case *Map:
		return h, true
	case map[string]any:
		m := NewMap()
		for _, k := range sortedStringKeys(h) {
			m.Set(k, h[k])
		}
		return m, true
	case map[Symbol]any:
		m := NewMap()
		keys := make([]string, 0, len(h))
		for k := range h {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		for _, k := range keys {
			m.Set(Symbol(k), h[Symbol(k)])
		}
		return m, true
	case map[any]any:
		m := NewMap()
		for k, val := range h {
			m.Set(k, val)
		}
		return m, true
	}
	return nil, false
}

func sortedStringKeys(h map[string]any) []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// asBigInt is the canonical big-integer view of an integer value.
func asBigInt(v any) (*big.Int, bool) {
	switch n := v.(type) {
	case int:
		return big.NewInt(int64(n)), true
	case int32:
		return big.NewInt(int64(n)), true
	case int64:
		return big.NewInt(n), true
	case *big.Int:
		return n, true
	}
	return nil, false
}

// Callable is a host closure used as a constructor function or a default block.
type Callable = func(any) any

// inspect renders v the way Ruby's Object#inspect does, for error messages. It
// covers the value shapes the type system produces error messages about.
func inspect(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return rubyStringInspect(x)
	case Symbol:
		return ":" + string(x)
	case int:
		return fmt.Sprintf("%d", x)
	case int32:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	case *big.Int:
		return x.String()
	case float64:
		return formatFloat(x)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = inspect(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case Date:
		return "#<Date: " + x.String() + ">"
	case time.Time:
		return x.String()
	}
	return fmt.Sprintf("%v", v)
}

// rubyStringInspect renders a Go string the way Ruby's String#inspect does for
// the characters that appear in the type system's error corpus.
func rubyStringInspect(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// formatFloat renders a float64 the way Ruby's Float#to_s does for the values
// the coercions produce (always at least one fractional digit).
func formatFloat(f float64) string {
	s := fmt.Sprintf("%g", f)
	if !strings.ContainsAny(s, ".eEnN") {
		s += ".0"
	}
	return s
}
