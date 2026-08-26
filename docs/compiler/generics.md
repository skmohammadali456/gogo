# Step 15: Generics

GOGO Step 15 adds generics to the existing single lexer, parser, semantic AST, grammar abstraction, canonical type system, and compiler/type-analysis pipeline.

## Syntax

Generic parameters are declared after a function, type alias, interface, or component name:

```gogo
create function identity<T>(x as T) as T { return x }
create function named<T extends Object{name: String}>(x as T) as String { return x.name }
create type Box<T> as Object{value: T}
create component Card<T>(value as "demo") { }
```

Type arguments use the same angle-bracket syntax in annotations and calls:

```gogo
let n as Number as identity<Number>(1)
let b as Box<String> as {value: "ok"}
```

Bengali and Hindi use the same punctuation and their existing vocabulary words for declarations, `as`, components, interfaces, and `extends`; no language-specific parser branch is introduced.

## Canonical model and scope

The canonical type package represents type parameters as `TypeParamKind` values with declaration-scoped stable IDs such as `func identity/T` or `type Box/T`. The ID is part of equality, so different generic declarations cannot accidentally share a parameter even when both are named `T`.

Generic aliases and named generic instantiations are resolved by the compiler session. The resolver substitutes canonical type parameters recursively through arrays, maps, sets, tuples, records/objects, `Optional`, unions, intersections, `Result`, enum payloads, and generic instances. Substitution is depth-bounded to terminate recursive/cyclic declarations safely.

## Constraints

Constraints use existing GOGO types and assignability. `T extends Object{name: String}` means an explicit or inferred type argument must be assignable to that object type. There is no separate constraint language.

## Inference

Function-call inference unifies declared parameter types with argument types. Contextual return type is used when available. Explicit type arguments take precedence over inference and are still checked against constraints. If a type parameter cannot be inferred, the call fails with a diagnostic rather than selecting an arbitrary type.

## Components and runtime boundary

Generic components are supported at the declaration/type-analysis level only. Step 15 does not add a UI framework, component engine, interpreter, class system, module system, decorators, formatter, or exceptions. Generics are erased at runtime boundaries: they exist as canonical instantiated types for analysis, diagnostics, equality, and assignability, but no runtime type objects are emitted or interpreted by this step.

## Defaults and limitations

Generic defaults are not implemented because the current README and grammar do not require a default type-argument syntax. Missing inference or constraint violations are therefore explicit errors. Nested closing generic brackets are recovered by the parser in type contexts without changing expression shift-token behavior.

## Diagnostics

Step 15 adds stable generic diagnostics for malformed generic parameter lists, duplicate parameters, invalid type arguments, failed inference, constraint violations, and non-terminating instantiation. Diagnostics render through the existing English, Bengali, and Hindi catalog and JSON renderer.
