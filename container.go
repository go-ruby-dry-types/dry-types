// Copyright (c) the go-ruby-dry-types/dry-types authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drytypes

// ArrayOf returns a member-typed array type (dry-types' `Array.of(T)`): it first
// requires an Array, then coerces every element through elem. The first element
// failure surfaces that element's error (matching the gem).
func ArrayOf(elem Type) *baseType {
	member := asBase(elem).fn
	return newType(func(v any) (any, error) {
		arr, ok := v.([]any)
		if !ok {
			return nil, constraintErr(v, "type?(Array, "+inspect(v)+") failed")
		}
		out := make([]any, len(arr))
		for i, e := range arr {
			c, err := member(e)
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	})
}

// SchemaKey is one member of a [HashSchema]: a key, its type, and whether the key
// is optional (dry-types' trailing `?` on the key name).
type SchemaKey struct {
	Key      Symbol
	Type     Type
	Optional bool
}

// HashSchema is a struct-hash type (dry-types' `Hash.schema({...})`). It embeds
// [*baseType], so it *is* a [Type]: it composes with every combinator and with
// [ArrayOf]. Build one with [NewSchema] and refine with [HashSchema.Strict].
type HashSchema struct {
	*baseType
	keys   []SchemaKey
	strict bool
}

// NewSchema builds a Hash.schema type from its members.
func NewSchema(keys ...SchemaKey) *HashSchema {
	s := &HashSchema{keys: keys}
	s.baseType = newType(s.coerce)
	return s
}

// Strict returns a copy of the schema that rejects unexpected keys
// (dry-types' `.strict`), raising [*UnknownKeysError].
func (s *HashSchema) Strict() *HashSchema {
	cp := &HashSchema{keys: s.keys, strict: true}
	cp.baseType = newType(cp.coerce)
	return cp
}

// AsType exposes the schema as a plain [Type] (identity — a *HashSchema already
// is a Type; kept for readability at call sites).
func (s *HashSchema) AsType() Type { return s.baseType }

func (s *HashSchema) coerce(v any) (any, error) {
	m, ok := asMap(v)
	if !ok {
		return nil, constraintErr(v, "type?(Hash, "+inspect(v)+") failed")
	}
	known := map[Symbol]SchemaKey{}
	for _, k := range s.keys {
		known[k.Key] = k
	}
	if s.strict {
		var unknown []Symbol
		for _, p := range m.Pairs() {
			if sym, isSym := keyToSymbol(p.Key); isSym {
				if _, ok := known[sym]; !ok {
					unknown = append(unknown, sym)
				}
			}
		}
		if len(unknown) > 0 {
			return nil, &UnknownKeysError{Message: "unexpected keys " + keyList(unknown) + " in Hash input"}
		}
	}
	out := NewMap()
	for _, k := range s.keys {
		val, present := lookupKey(m, k.Key)
		if !present {
			if k.Optional {
				continue
			}
			return nil, &MissingKeyError{Message: ":" + string(k.Key) + " is missing in Hash input"}
		}
		c, err := asBase(k.Type).Call(val)
		if err != nil {
			return nil, wrapSchemaErr(k.Key, val, err)
		}
		out.Set(k.Key, c)
	}
	// dry-types' default schema drops keys that are not declared members (only a
	// `.strict` schema errors on them; both discard undeclared keys from output).
	return out, nil
}

// wrapSchemaErr turns a member failure into the gem's SchemaError shape,
// `<val> (<Class>) has invalid type for :key <inner rule>`, reusing the inner
// constraint rule when present.
func wrapSchemaErr(key Symbol, val any, err error) error {
	var rule string
	if ce, ok := err.(*ConstraintError); ok {
		rule = ce.Rule
	} else {
		// A coercion failure inside a schema member reports its message as the
		// rule with a trailing " failed" (dry-types wraps the coercion result).
		rule = err.Error() + " failed"
	}
	msg := inspect(val) + " (" + valueClass(val) + ") has invalid type for :" + string(key) +
		" violates constraints (" + rule + ")"
	return &SchemaError{Message: msg}
}

// keyToSymbol normalizes a hash key to a Symbol for schema matching (schema keys
// are symbols; string keys are accepted and compared by name).
func keyToSymbol(k any) (Symbol, bool) {
	switch x := k.(type) {
	case Symbol:
		return x, true
	case string:
		return Symbol(x), true
	}
	return "", false
}

// lookupKey finds a schema key in the input map, accepting either a Symbol or a
// String key of the same name.
func lookupKey(m *Map, key Symbol) (any, bool) {
	if v, ok := m.Get(key); ok {
		return v, true
	}
	if v, ok := m.Get(string(key)); ok {
		return v, true
	}
	return nil, false
}
