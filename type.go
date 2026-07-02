// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// Undefined is dry-types' sentinel for "no value given" — it triggers a
// [Type.Default]'s fallback. Pass it as the input to Call to request the
// default. (In the gem this is Dry::Types::Undefined.)
var Undefined = undefined{}

type undefined struct{}

// Type is a composable dry-types type: it coerces-and-validates an input, and
// carries the combinator methods so types chain fluently (e.g.
// `StrictInteger().Constrained(...).Optional()`).
//
// Call applies the type: it returns the coerced value on success or a
// [*CoercionError] / [*ConstraintError] / schema error on failure — the same
// value and message the dry-types gem's `type[input]` produces.
//
// The interface is closed to this package (every Type is a *baseType) via the
// unexported node method, which keeps combinator composition total.
type Type interface {
	// Call coerces and validates input, returning the coerced value or an error.
	Call(input any) (any, error)

	// Optional returns a type that also accepts nil (dry-types' `.optional`).
	Optional() Type
	// Default substitutes val when the input is [Undefined] (`.default(val)`).
	Default(val any) Type
	// DefaultFn substitutes fn() when the input is [Undefined] (`.default { }`).
	DefaultFn(fn func() any) Type
	// Constructor pre-processes input through fn, then applies the base type.
	Constructor(fn Callable) Type
	// Meta returns a copy carrying the merged metadata (`.meta(...)`).
	Meta(m map[string]any) Type
	// GetMeta returns the attached metadata.
	GetMeta() map[string]any
	// Or returns the sum type `self | other` (`A | B`).
	Or(other Type) Type
	// Enum requires the coerced value to be one of values (`.enum(...)`).
	Enum(values ...any) Type
	// Constrained applies dry-logic predicates after coercion (`.constrained`).
	Constrained(cs ...Constraint) Type

	// node is the internal upcast to the single concrete implementation.
	node() *baseType
}

// Valid reports whether t accepts input (dry-types' `type.valid?`).
func Valid(t Type, input any) bool {
	_, err := t.Call(input)
	return err == nil
}

// Try applies t and returns a [Result] (dry-types' `type.try`): Success carries
// the coerced value, Failure carries the error.
func Try(t Type, input any) Result {
	out, err := t.Call(input)
	if err != nil {
		return Result{success: false, input: input, err: err}
	}
	return Result{success: true, input: out}
}

// Result is the outcome of [Try]: a success carrying the coerced value or a
// failure carrying the error (mirrors Dry::Types::Result::Success/Failure).
type Result struct {
	success bool
	input   any
	err     error
}

// Success reports whether the Result is a success.
func (r Result) Success() bool { return r.success }

// Failure reports whether the Result is a failure.
func (r Result) Failure() bool { return !r.success }

// Input returns the coerced value (on success) or the offending input (on
// failure), matching Dry::Types::Result#input.
func (r Result) Input() any { return r.input }

// Error returns the failure's error, or nil on success.
func (r Result) Error() error { return r.err }
