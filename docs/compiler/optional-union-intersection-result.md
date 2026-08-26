# Step 12: Optional, union, intersection, and result types

Step 12 extends the single canonical `internal/types.Type` model; it does not add a second type system, language-specific AST, or alternate compiler pipeline. Surface annotations still flow through `ast.TypeRef`, parser `parseType`, and compiler `ResolveType`.

## Syntax

The shared grammar accepts the following type forms in English, Bengali, and Hindi sessions because only declaration keywords are localized:

- `Optional<T>` for explicit absent-or-present values.
- `A | B` for unions. Unions are parsed as type expressions and may contain primitives, literals, arrays, maps, sets, tuples, objects, aliases, and results.
- `A & B` for intersections. Intersections bind tighter than unions.
- `Result<Ok, Err>` for recoverable success or failure.
- Parenthesized type expressions such as `(String | Number)[]`.

No Step 13 narrowing, matching, inference, classes, interfaces, modules, or exception system is introduced.

## Canonical normalization and equality

Normalization recursively normalizes nested members and then canonicalizes composite types. Unions and intersections flatten nested expressions of the same kind, sort by canonical string form, remove duplicate equivalent members, and remove redundant literal/structural members when a stricter or broader member already determines the same value space. Single-member unions and intersections collapse to that member. Equality is structural and deterministic; `A | B` equals `B | A`, and `A & B` equals `B & A` after normalization. `Optional<T>` is not equal to `T`, and nested optionals remain explicit.

Alias names resolve before equality or assignability is checked, and aliases remain session-local. Cyclic aliases are rejected through existing type diagnostics.

## Assignability

Assignability is intentionally stricter than equality:

- `Optional<T>` is assignable only to compatible `Optional<U>`; `T` is not silently assignable to `Optional<T>` and `Optional<T>` is not assignable to `T`.
- A concrete source is assignable to `A | B` when it satisfies at least one member.
- A union source is assignable to a target only when every possible member is assignable to that target.
- A source is assignable to `A & B` only when it satisfies every member.
- An intersection source is assignable to a target if one of its required member views is assignable to the target.
- `Result<Ok, Err>` is assignable to `Result<Ok2, Err2>` when both payload types are assignable in their respective positions.

## Objects, readonly, optional properties, and conflicts

Object intersections merge compatible fields structurally. Required beats optional because an intersection must satisfy all requirements; readonly is preserved if either side is readonly. Compatible string index signatures are preserved, and named fields must satisfy any preserved index signature. Conflicting object property or index-signature types produce a non-collapsed intersection that no ordinary conflicting object can satisfy. Object assignability retains Step 11 width-subtyping, optional-property, readonly-property, and index-signature checks.

## Runtime values and mutability

`types.Value` remains the defensive-copy runtime boundary. Optional runtime values are explicit absent or present wrappers. Union values store the canonical active member accepted for the payload. Intersection values must satisfy all member requirements. Result values carry an `OK` flag and validate payloads against the `Ok` or `Err` parameter. Optional mutability follows the payload; union and intersection mutability are mutable if any member is mutable; result mutability is mutable if either payload is mutable.

## Diagnostics

Malformed Step 12 annotations and incompatible assignments use the existing Step 4 diagnostic pipeline, preserving source spans and UTF-8 positions. Stable Step 12 catalog entries reserve `G3005` for malformed optional/union/intersection/result types and `G3006` for invalid composite assignability, while existing `G3001` and `G3002` continue to report annotation-resolution and assignment failures at call sites.
