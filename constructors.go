// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// typeCheck builds the `type?` predicate a strict/enum-style type applies: it
// returns the value unchanged on a class match, else a ConstraintError shaped
// `<v> violates constraints (type?(<Class>, <v>) failed)`.
func typeCheck(p primitive) func(any) (any, error) {
	return func(v any) (any, error) {
		if p.matches(v) {
			return v, nil
		}
		return nil, constraintErr(v, "type?("+p.className()+", "+inspect(v)+") failed")
	}
}

// strict builds a Strict::<T> type: pure type? check, no coercion.
func strict(p primitive) *baseType { return newType(typeCheck(p)) }

// nominalOf builds a Nominal::<T> type: no constraint at all, passes input
// through unchanged (dry-types' Nominal is the unconstrained base).
func nominalOf(_ primitive) *baseType {
	return newType(func(v any) (any, error) { return v, nil })
}

// coercibleWith builds a Coercible/Params/JSON type: run coerce, then (for the
// strict-typed variants) the value is already the right class so no extra check
// is applied — a coercion failure surfaces its CoercionError directly.
func coercibleWith(coerce func(any) (any, error)) *baseType {
	return newType(coerce)
}

// boolStrict builds Strict::Bool = TrueClass | FalseClass. The gem models it as a
// sum; a non-bool reports the second branch (FalseClass) in its message, which we
// reproduce.
func boolStrict() *baseType {
	return newType(func(v any) (any, error) {
		if _, ok := v.(bool); ok {
			return v, nil
		}
		return nil, constraintErr(v, "type?(FalseClass, "+inspect(v)+") failed")
	})
}

// --- Strict namespace ---

// StrictInteger is Types::Strict::Integer.
func StrictInteger() Type { return strict(pInteger) }

// StrictFloat is Types::Strict::Float.
func StrictFloat() Type { return strict(pFloat) }

// StrictString is Types::Strict::String.
func StrictString() Type { return strict(pString) }

// StrictSymbol is Types::Strict::Symbol.
func StrictSymbol() Type { return strict(pSymbol) }

// StrictBool is Types::Strict::Bool.
func StrictBool() Type { return boolStrict() }

// StrictNil is Types::Strict::Nil.
func StrictNil() Type { return strict(pNil) }

// StrictArray is Types::Strict::Array.
func StrictArray() Type { return strict(pArray) }

// StrictHash is Types::Strict::Hash.
func StrictHash() Type { return strict(pHash) }

// StrictDate is Types::Strict::Date.
func StrictDate() Type { return strict(pDate) }

// StrictTime is Types::Strict::Time.
func StrictTime() Type { return strict(pTime) }

// StrictDateTime is Types::Strict::DateTime.
func StrictDateTime() Type { return strict(pDateTime) }

// --- Nominal namespace (unconstrained) ---

// NominalInteger is Types::Nominal::Integer (no constraint).
func NominalInteger() Type { return nominalOf(pInteger) }

// NominalString is Types::Nominal::String (no constraint).
func NominalString() Type { return nominalOf(pString) }

// NominalFloat is Types::Nominal::Float (no constraint).
func NominalFloat() Type { return nominalOf(pFloat) }

// NominalSymbol is Types::Nominal::Symbol (no constraint).
func NominalSymbol() Type { return nominalOf(pSymbol) }

// NominalBool is Types::Nominal::Bool (no constraint).
func NominalBool() Type { return nominalOf(pTrue) }

// NominalArray is Types::Nominal::Array (no constraint).
func NominalArray() Type { return nominalOf(pArray) }

// NominalHash is Types::Nominal::Hash (no constraint).
func NominalHash() Type { return nominalOf(pHash) }

// --- Coercible namespace ---

// CoercibleInteger is Types::Coercible::Integer (Kernel#Integer).
func CoercibleInteger() Type { return coercibleWith(kernelInteger) }

// CoercibleFloat is Types::Coercible::Float (Kernel#Float).
func CoercibleFloat() Type { return coercibleWith(kernelFloat) }

// CoercibleString is Types::Coercible::String (#to_s).
func CoercibleString() Type { return coercibleWith(kernelString) }

// CoercibleSymbol is Types::Coercible::Symbol (#to_sym).
func CoercibleSymbol() Type { return coercibleWith(kernelSymbol) }

// --- Params namespace (form-param coercion) ---

// ParamsInteger is Types::Params::Integer.
func ParamsInteger() Type { return coercibleWith(paramsInteger) }

// ParamsFloat is Types::Params::Float.
func ParamsFloat() Type { return coercibleWith(paramsFloat) }

// ParamsBool is Types::Params::Bool.
func ParamsBool() Type { return coercibleWith(paramsBool) }

// ParamsNil is Types::Params::Nil ("" → nil).
func ParamsNil() Type { return coercibleWith(paramsNil) }

// ParamsSymbol is Types::Params::Symbol.
func ParamsSymbol() Type { return coercibleWith(paramsSymbol) }

// ParamsDate is Types::Params::Date.
func ParamsDate() Type { return coercibleWith(coerceDate) }

// ParamsTime is Types::Params::Time.
func ParamsTime() Type { return coercibleWith(coerceTime) }

// ParamsDateTime is Types::Params::DateTime.
func ParamsDateTime() Type { return coercibleWith(coerceDateTime) }

// --- JSON namespace ---

// JSONDate is Types::JSON::Date.
func JSONDate() Type { return coercibleWith(coerceDate) }

// JSONTime is Types::JSON::Time.
func JSONTime() Type { return coercibleWith(coerceTime) }

// JSONDateTime is Types::JSON::DateTime.
func JSONDateTime() Type { return coercibleWith(coerceDateTime) }

// JSONSymbol is Types::JSON::Symbol.
func JSONSymbol() Type { return coercibleWith(kernelSymbol) }

// JSONNil is Types::JSON::Nil (passes nil through, else type-checks nil).
func JSONNil() Type { return strict(pNil) }
