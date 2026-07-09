# frozen_string_literal: true
#
# Basic usage of dry-types from Ruby, running under go-embedded-ruby (rbgo).
#
# A type is a value: look one up with `Dry::Types[...]`, then apply it to an
# input to coerce and validate, or refine it with combinator methods.

require "dry-types"

# Strict types check the input's class and pass it through unchanged.
p Dry::Types["strict.integer"].call(3)          # => 3

# Coercible types apply the target's Ruby coercion rules.
p Dry::Types["coercible.integer"].call("7")     # => 7

# Params types coerce form-style strings ("true" -> true, "" -> nil).
p Dry::Types["params.bool"].call("true")        # => true

# Non-raising checks: valid? and try.
p Dry::Types["strict.integer"].valid?("x")      # => false
p Dry::Types["strict.integer"].try("x").success? # => false

# Combinators build new types from existing ones.
p Dry::Types["strict.integer"].optional.call(nil)          # => nil
p Dry::Types["strict.integer"].constrained(gt: 5).call(6)  # => 6
p Dry::Types["strict.string"].enum("a", "b").call("a")     # => "a"

# Sum types accept either branch.
either = Dry::Types["strict.integer"] | Dry::Types["strict.string"]
p either.call("x")                              # => "x"

# Array and Hash schemas coerce each element / member.
p Dry::Types.ArrayOf(Dry::Types["coercible.integer"]).call(["1", "2"]) # => [1, 2]
p Dry::Types.Schema(name: Dry::Types["strict.string"]).call({name: "x"}) # => {name: "x"}

# A failed constraint raises with the gem's exact message.
begin
  Dry::Types["strict.integer"].constrained(gt: 5).call(1)
rescue Dry::Types::ConstraintError => e
  puts e.message                                # => 1 violates constraints (gt?(5, 1) failed)
end
