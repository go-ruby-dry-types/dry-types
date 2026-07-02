<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-dry-types/brand/main/social/go-ruby-dry-types-dry-types.png" alt="go-ruby-dry-types/dry-types" width="720"></p>

# dry-types — go-ruby-dry-types

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-dry-types.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) MRI-faithful reimplementation of Ruby's
[`dry-types`](https://dry-rb.org/gems/dry-types/) gem** — a composable type
system with coercion, constraints, and combinators. Every type is a value you
build from a constructor (`Strict`, `Coercible`, `Params`, `JSON`, `Nominal`) and
refine with combinator methods (`Optional`, `Default`, `Constrained`, `Enum`,
`Or`, `Constructor`, `Array.of`, `Hash.schema`). Applying a type to an input
coerces and validates it, returning the coerced value or an error whose message
is **byte-identical** to the `dry-types` gem's — **without any Ruby runtime**.

It is the type-system backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) and feeds the
`dry-struct` / `dry-validation` ports; it is a **standalone, reusable** module
with no dependency on the Ruby runtime.

## Features

Faithful port of `dry-types`' coercion + validation, differentially validated
against the `dry-types` gem on Ruby ≥ 4.0:

- **Nominal / strict types** — `Strict::{Integer,String,Float,Bool,Symbol,Array,
  Hash,Date,Time,DateTime,Nil}` (pure `type?` check, raise `ConstraintError`),
  and unconstrained `Nominal::*` pass-through types.
- **Coercible types** — `Coercible::{Integer,Float,String,Symbol}` using the
  target's Ruby coercion rules (`Kernel#Integer` with `0x`/`0o`/`0b` prefixes and
  underscore separators, `Kernel#Float`, `#to_s`, `#to_sym`); raise
  `CoercionError`.
- **Params types** — form-param coercion: `Params::{Integer,Float,Bool,Nil,
  Symbol,Date,Time,DateTime}`, with `"1"`→`1`, `"true"`→`true`, `""`→`nil`, and
  the gem's exact `TRUE_VALUES` / `FALSE_VALUES` sets.
- **JSON types** — `JSON::{Date,Time,DateTime,Symbol,Nil}`.
- **Combinators** — `.optional` (nil-allowed), `.default(x)` and callable
  defaults, `.constrained(gt:, gteq:, lt:, lteq:, format:, size:, min_size:,
  max_size:, included_in:, excluded_from:, filled:, eql:, …)`, `.enum(...)`, sum
  types `A | B`, `.meta`, `.constructor(fn)`, `Array.of(T)`, and `Hash.schema`
  with required / optional keys and `.strict`.
- **Exact error parity** — `CoercionError`, `ConstraintError`, `SchemaError`,
  `MissingKeyError`, `UnknownKeysError` messages match the gem verbatim.

## Usage

```go
import drytypes "github.com/go-ruby-dry-types/dry-types"

age := drytypes.CoercibleInteger()
v, err := age.Call("30")            // v == int64(30)

adult := drytypes.StrictInteger().Constrained(
    drytypes.Constraint{Name: "gteq", Arg: 18})
_, err = adult.Call(17)             // "17 violates constraints (gteq?(18, 17) failed)"

status := drytypes.StrictString().Enum("draft", "published")
_, err = status.Call("x")           // included_in?(...) constraint error

user := drytypes.NewSchema(
    drytypes.SchemaKey{Key: "name", Type: drytypes.StrictString()},
    drytypes.SchemaKey{Key: "age", Type: drytypes.CoercibleInteger()},
)
out, err := user.Call(map[string]any{"name": "Jane", "age": "30"})
```

Every type also answers `drytypes.Valid(t, input) bool` and
`drytypes.Try(t, input) Result` (a Success / Failure carrying the coerced value
or the error).

## Value model

Ruby values are the same small, fixed set of Go types the `go-ruby-*` ecosystem
uses, so a host (`go-embedded-ruby` / `rbgo`) maps its object graph to and from
this package with no glue: `nil`, `bool`, `int64` / `*big.Int`, `float64`,
`string`, `Symbol`, `[]any`, `*Map` (ordered hash), `Date`, and `Time`. Host
callables (default blocks, constructor functions) are Go closures.

## Tests & coverage

The deterministic, ruby-free suite holds **100% statement coverage** on its own.
A differential **oracle** additionally compares every coercion, constraint, and
combinator outcome — success values *and* error messages — against the real
`dry-types` gem (gated on `RUBY_VERSION >= "4.0"`); it skips itself where `ruby`
or the gem is absent (the qemu cross-arch lanes and the Windows lane), so the
gate stays green everywhere. Validated on all six 64-bit Go targets
(`amd64`, `arm64`, `riscv64`, `loong64`, `ppc64le`, `s390x`) across Linux, macOS,
and Windows.

```sh
go test -race -cover ./...
```

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) 2026, the
go-ruby-dry-types/dry-types authors.
