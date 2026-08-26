# Step 10 canonical type system

## Architecture and boundary

`internal/types` is the single canonical language-independent type model. It has no dependency on lexer tokens, AST nodes, grammar vocabulary, compiler sessions, or Go runtime containers. `ast.TypeRef` remains a surface annotation and `compiler.ResolveType` is the only annotation-to-type conversion path. This preserves the existing one lexer, parser, AST, grammar layer, compiler pipeline, diagnostics system, and project configuration system.

The Step 10 runtime boundary is `types.Value { Type, Data }`: it preserves canonical shape while deliberately not selecting an execution engine. A later runtime/IR step owns execution, storage layout, collection algorithms, and backend conversion.

## Types and representation

Canonical primitive singleton values are `string`, `number`, `boolean`, `bigint`, and `bytes`. `array<T>`, `map<K,V>`, and `set<T>` carry their element information and are mutable collection values. `tuple<T...>` has ordered positional members and is immutable. `record{field: T...}` is an immutable, closed named aggregate. Record fields are sorted by name at construction: declaration order is not semantic, duplicate names are rejected, and equality is deterministic. All constructors defensively copy slices/maps exposed through accessors.

Nested structures recurse naturally: `array<string>`, `map<string, number>`, `set<boolean>`, `tuple<string, number>`, and records containing collections retain each component type.

## Equality, assignability, literals, and mutability

Equality is structural and deterministic, including collection members recursively. Map key/value and tuple position must match. Record equality matches field names and field types after name sorting. Mutability is part of equality, so mutable and immutable variants do not compare equal.

Assignability is intentionally narrower than equality: equal types assign; a literal type is assignable to its primitive base but is not equal to it; all Step 10 collections are invariant; tuples require the same length and position types; records require the same closed field set and types. No Go compatibility rules are used.

A literal records its primitive base plus exact canonical source text. The canonical model can represent string, number, boolean, bigint, and bytes literals as distinct literal types, while each is assignable to its matching primitive. The current surface lexer intentionally provides only its pre-existing string, number, and bigint forms; boolean and bytes literal syntax is deferred until a localized grammar/token contract is specified. Primitives and tuples/records are immutable values. Arrays, maps, and sets are mutable values. Existing `constant` and `variable` declarations now preserve binding mutability and direct identifier reassignment is checked; collection mutation syntax beyond the current declaration/assignment syntax remains deliberately outside Step 10.

## Surface syntax and multilingual behavior

Existing `TypeRef` syntax now accepts `Array<String>`, `Map<String, Number>`, `Set<Boolean>`, `Tuple<String, Number>`, `Record{name: String}`, plus the existing `Text[]` suffix. Core spelling resolution happens after parsing; grammar sessions still supply only localized declaration vocabulary. English, Bengali, and Hindi therefore produce the same canonical types, and Unicode identifiers remain source data rather than type identity.

## Diagnostics

The existing diagnostics bag reports `G3001` for unsupported/malformed canonical annotations and `G3002` for a known literal/collection value incompatible with a declaration. Diagnostics retain AST source spans and use the existing renderer/localization fallback mechanism.
