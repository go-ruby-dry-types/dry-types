// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

import (
	"math/big"
	"time"
)

// primitive names a Ruby primitive class a strict/nominal type checks against.
type primitive int

const (
	pInteger primitive = iota
	pFloat
	pString
	pSymbol
	pTrue  // TrueClass
	pFalse // FalseClass
	pNil
	pArray
	pHash
	pDate
	pTime
	pDateTime
)

// className is the Ruby class name used in a `type?(<Class>, <val>)` message.
func (p primitive) className() string {
	switch p {
	case pInteger:
		return "Integer"
	case pFloat:
		return "Float"
	case pString:
		return "String"
	case pSymbol:
		return "Symbol"
	case pTrue:
		return "TrueClass"
	case pFalse:
		return "FalseClass"
	case pNil:
		return "NilClass"
	case pArray:
		return "Array"
	case pHash:
		return "Hash"
	case pDate:
		return "Date"
	case pTime:
		return "Time"
	case pDateTime:
		return "DateTime"
	}
	return "Object"
}

// matches reports whether v is an instance of the primitive class.
func (p primitive) matches(v any) bool {
	switch p {
	case pInteger:
		switch v.(type) {
		case int, int32, int64, *big.Int:
			return true
		}
	case pFloat:
		_, ok := v.(float64)
		return ok
	case pString:
		_, ok := v.(string)
		return ok
	case pSymbol:
		_, ok := v.(Symbol)
		return ok
	case pTrue:
		b, ok := v.(bool)
		return ok && b
	case pFalse:
		b, ok := v.(bool)
		return ok && !b
	case pNil:
		return v == nil
	case pArray:
		_, ok := v.([]any)
		return ok
	case pHash:
		switch v.(type) {
		case *Map, map[string]any, map[Symbol]any, map[any]any:
			return true
		}
	case pDate:
		_, ok := v.(Date)
		return ok
	case pTime, pDateTime:
		_, ok := v.(time.Time)
		return ok
	}
	return false
}
