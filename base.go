// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// baseType is the single concrete Type. Every constructor and combinator returns
// a *baseType; combinators build a new one that wraps the receiver's fn. This
// mirrors dry-types' decorator design (Constrained/Default/Sum/… wrap a Nominal)
// while staying a closed, total, allocation-light Go value.
type baseType struct {
	// fn coerces-and-validates. It is the composed pipeline.
	fn func(input any) (any, error)
	// meta is the attached metadata map (dry-types' .meta).
	meta map[string]any
}

func (b *baseType) node() *baseType { return b }

// Call runs the pipeline.
func (b *baseType) Call(input any) (any, error) { return b.fn(input) }

// newType wraps a coerce-and-validate function in a *baseType.
func newType(fn func(any) (any, error)) *baseType {
	return &baseType{fn: fn}
}

// derive returns a copy of b with a new fn, carrying meta forward.
func (b *baseType) derive(fn func(any) (any, error)) *baseType {
	return &baseType{fn: fn, meta: b.meta}
}

// asBase is the internal upcast used by combinators that take a Type argument.
func asBase(t Type) *baseType { return t.node() }
