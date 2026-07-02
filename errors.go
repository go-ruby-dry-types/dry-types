// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// CoercionError is raised when a coercion fails (e.g. Integer("abc")). Its
// message is byte-identical to Dry::Types::CoercionError#message.
type CoercionError struct{ Message string }

func (e *CoercionError) Error() string { return e.Message }

func coercionErr(msg string) *CoercionError { return &CoercionError{Message: msg} }

// ConstraintError is raised when a value violates a constraint (including the
// implicit type? constraint of a strict type). Its message matches
// Dry::Types::ConstraintError#message: `<input> violates constraints (<rule>)`.
type ConstraintError struct {
	Message string
	// Input is the offending value.
	Input any
	// Rule is the failed predicate rendered as dry-logic renders it
	// (e.g. `gt?(18, 10) failed`).
	Rule string
}

func (e *ConstraintError) Error() string { return e.Message }

func constraintErr(input any, rule string) *ConstraintError {
	return &ConstraintError{
		Message: inspect(input) + " violates constraints (" + rule + ")",
		Input:   input,
		Rule:    rule,
	}
}

// MissingKeyError is raised by a strict schema when a required key is absent:
// `:key is missing in Hash input`.
type MissingKeyError struct{ Message string }

func (e *MissingKeyError) Error() string { return e.Message }

// UnknownKeysError is raised by a strict schema on unexpected keys:
// `unexpected keys [:a, :b] in Hash input`.
type UnknownKeysError struct{ Message string }

func (e *UnknownKeysError) Error() string { return e.Message }

// SchemaError is raised when a schema member fails coercion/validation:
// `<value> (<Class>) has invalid type for :key <inner rule>`.
type SchemaError struct{ Message string }

func (e *SchemaError) Error() string { return e.Message }
