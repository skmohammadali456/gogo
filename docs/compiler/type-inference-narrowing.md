# Step 13: type inference and narrowing

Step 13 extends the existing compiler type-analysis pass; it does not add a second parser, AST, grammar, compiler pipeline, or type system.

## Inference model

Each binding tracks three distinct facts:

- **Declared type**: the authoritative annotation, or the normalized variable type inferred at declaration time when no annotation is present.
- **Inferred type**: the expression-local type inferred from literals, arrays, objects, member access, calls, conditionals, and existing bindings.
- **Contextual type**: a target type pushed into array/object/function-call arguments and annotated initializers so literals and object members can be checked against aliases, unions, optionals, and result payloads.

Aliases are resolved through the existing session resolver and all analysis uses `internal/types.Type` canonical values.

## Narrowing and truthiness

Control-flow analysis copies a lexical environment for each branch and updates only the branch-local current type. The declared type is never mutated globally. At joins the analysis restores the declared type unless an early return makes exactly one branch reachable.

Defined truthiness is intentionally limited to `Boolean`, `Optional<T>`, `Result<Ok, Err>`, and unions involved in explicit proofs. `Optional<T>` narrows to `T` in a proven presence branch. `Result<Ok, Err>` narrows to `Ok` in the truthy branch and `Err` in the falsy branch. GOGO does not use JavaScript-style truthiness for arbitrary strings, numbers, arrays, or objects.

## Property checks and discriminated unions

Property existence checks and member equality checks narrow object unions by selecting variants that prove the property exists or whose property type is compatible with the checked literal. A common literal property such as `kind` acts as a discriminant without introducing classes or interfaces. Nested property reads use the ordinary semantic AST and canonical object fields.

## Exhaustiveness

Step 13 documents exhaustiveness for supported proof forms. The current syntax has `if`/`else`; an `else` paired with a successful discriminant check covers the remaining members. Dedicated match syntax is intentionally not added because that belongs to a future step.

## Function and UI-state boundaries

Annotated function parameters seed local bindings, annotated return expressions are checked where the current architecture can infer them, and branch-local narrowing does not escape a function. UI state is modeled only as ordinary canonical object, union, optional, and result types; no UI-specific engine or global mutable analysis state is introduced.
