// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// Optional returns a type that also accepts nil (dry-types' `.optional`, i.e.
// `Nil | self`). A nil input yields nil; any other input goes through the base.
func (b *baseType) Optional() Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		if v == nil {
			return nil, nil
		}
		return base(v)
	})
}

// Default returns a type that substitutes val when the input is [Undefined]
// (dry-types' `.default(val)`). The substituted value is returned as-is (the gem
// does not re-run coercion on a static default).
func (b *baseType) Default(val any) Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		if _, ok := v.(undefined); ok {
			return val, nil
		}
		return base(v)
	})
}

// DefaultFn returns a type whose default is produced by calling fn (dry-types'
// `.default { ... }` block form).
func (b *baseType) DefaultFn(fn func() any) Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		if _, ok := v.(undefined); ok {
			return fn(), nil
		}
		return base(v)
	})
}

// Constructor returns a type that first runs fn over the input, then applies the
// base type to fn's result (dry-types' `.constructor(fn)`).
func (b *baseType) Constructor(fn Callable) Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		return base(fn(v))
	})
}

// Meta returns a copy of the type carrying the merged metadata (dry-types'
// `.meta(...)`).
func (b *baseType) Meta(m map[string]any) Type {
	merged := map[string]any{}
	for k, v := range b.meta {
		merged[k] = v
	}
	for k, v := range m {
		merged[k] = v
	}
	return &baseType{fn: b.fn, meta: merged}
}

// GetMeta returns the attached metadata (dry-types' `.meta`).
func (b *baseType) GetMeta() map[string]any {
	if b.meta == nil {
		return map[string]any{}
	}
	return b.meta
}

// Or returns the sum type `self | other` (dry-types' `A | B`): it tries the base
// first, then other; on total failure it reports other's error (matching the
// gem, whose sum surfaces the right-hand branch's message).
func (b *baseType) Or(other Type) Type {
	left := b.fn
	right := asBase(other).fn
	return b.derive(func(v any) (any, error) {
		if out, err := left(v); err == nil {
			return out, nil
		}
		return right(v)
	})
}

// Enum returns a type that additionally requires the (coerced) value to be one of
// values (dry-types' `.enum(...)`), reporting the `included_in?` constraint on a
// miss.
func (b *baseType) Enum(values ...any) Type {
	base := b.fn
	return b.derive(func(v any) (any, error) {
		out, err := base(v)
		if err != nil {
			return nil, err
		}
		for _, want := range values {
			if valuesEqual(out, want) {
				return out, nil
			}
		}
		return nil, constraintErr(out, "included_in?("+arrayToS(values)+", "+inspect(out)+") failed")
	})
}
