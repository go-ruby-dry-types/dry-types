// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math/big"
	"testing"
	"time"
)

func TestPrimitiveClassName(t *testing.T) {
	cases := []struct {
		p    primitive
		want string
	}{
		{pInteger, "Integer"},
		{pFloat, "Float"},
		{pString, "String"},
		{pSymbol, "Symbol"},
		{pTrue, "TrueClass"},
		{pFalse, "FalseClass"},
		{pNil, "NilClass"},
		{pArray, "Array"},
		{pHash, "Hash"},
		{pDate, "Date"},
		{pTime, "Time"},
		{pDateTime, "DateTime"},
		{primitive(999), "Object"},
	}
	for _, c := range cases {
		if got := c.p.className(); got != c.want {
			t.Fatalf("className(%d)=%s want %s", c.p, got, c.want)
		}
	}
}

func TestPrimitiveMatches(t *testing.T) {
	now := time.Now()
	cases := []struct {
		p    primitive
		v    any
		want bool
	}{
		{pInteger, 5, true},
		{pInteger, int32(5), true},
		{pInteger, int64(5), true},
		{pInteger, big.NewInt(5), true},
		{pInteger, "5", false},
		{pFloat, 3.5, true},
		{pFloat, 3, false},
		{pString, "x", true},
		{pString, 5, false},
		{pSymbol, Symbol("s"), true},
		{pSymbol, "s", false},
		{pTrue, true, true},
		{pTrue, false, false},
		{pTrue, "x", false},
		{pFalse, false, true},
		{pFalse, true, false},
		{pFalse, "x", false},
		{pNil, nil, true},
		{pNil, 5, false},
		{pArray, []any{}, true},
		{pArray, 5, false},
		{pHash, NewMap(), true},
		{pHash, map[string]any{}, true},
		{pHash, map[Symbol]any{}, true},
		{pHash, map[any]any{}, true},
		{pHash, 5, false},
		{pDate, Date{}, true},
		{pDate, 5, false},
		{pTime, now, true},
		{pTime, 5, false},
		{pDateTime, now, true},
		{pDateTime, 5, false},
		{primitive(999), 5, false},
	}
	for _, c := range cases {
		if got := c.p.matches(c.v); got != c.want {
			t.Fatalf("matches(%d, %v)=%v want %v", c.p, c.v, got, c.want)
		}
	}
}
