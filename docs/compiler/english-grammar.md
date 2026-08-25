# English GOGO grammar

Step 6 makes English the primary surface grammar while preserving the Step 5 pipeline: source words are resolved by the active vocabulary into canonical grammar keywords, the parser consumes those symbols, and all surfaces produce the same semantic AST.

## Vocabulary

Reserved English keywords are `create`, `variable`, `constant`, `function`, `component`, `import`, `from`, `as`, `return`, `if`, and `else`. Concise aliases are `let` for `create variable`, `const` for `create constant`, and `fn` for `create function`. Type aliases are contextual in type positions: `Text`/`String`, `Number`, `Boolean`, `Array`, and `Object`.

## Statements and termination

Statements may end with `;`, `}`, EOF, or a newline before the next statement. Newlines are not tokens; the parser compares source line numbers, so multiline expressions continue when an operator, comma, bracket, parenthesis, or required delimiter keeps the statement open.

## Declarations

Readable variable declaration:

```gogo
create variable user as "Alex"
```

Concise equivalent:

```gogo
let user as "Alex"
```

Optional type annotation appears between the name and initializer:

```gogo
create variable user as Text as "Alex"
```

Constants use the same shape with `constant` or `const`. Assignments are expressions whose left side must be an identifier, member, or index.

## Control flow and functions

Conditionals use `if <expression> { ... }` with optional `else { ... }`. Functions use `create function name(parameters) { ... }` or `fn name(parameters) { ... }`. Parameters may be `name` or `name as Type`; return type may be `as Type` before the body. Return statements require a value.

## Types

Step 6 recognizes named types and array suffixes in type positions: `Text`, `Number`, `Boolean`, `Object`, `User`, and `Text[]`. Type parsing is contextual and never changes expression identifier parsing.

## Imports

Imports establish syntax only: `import "module"` and `import "module" as alias`.

## Components

Component declarations use `create component Name(prop as value) { ... }`. Properties use `name as expression`; children/content are ordinary statements inside the component block. Runtime UI behavior is intentionally not implemented in Step 6.

## Operators

From lowest to highest precedence: assignment (`=`, compound assignment, right associative), `||`/`??`, `&&`, `|`, `^`, `&`, equality, comparison and shifts, `+`/`-`, `*`/`/`/`%`, exponentiation (`**`, right associative), prefix unary (`!`, `+`, `-`, `~`, `++`, `--`), and postfix call/member/index. Conditional `?:` binds lower than binary expressions and above statement termination.

## Ambiguity rules

Keywords are reserved only when the active vocabulary maps them and the grammar expects that keyword; object property names and member names remain ordinary identifiers. `{` starts a block in statement position and an object literal in expression position. A `create` keyword dispatches by the following grammar symbol. Type names are parsed only after contextual `as` in declarations, parameters, and return annotations; `as` before a value initializer is chosen by one-token lookahead and expression viability. Newline terminates a statement only after a complete statement, not inside delimiter lists or after an operator. Component declarations require `create component` or its canonical keyword sequence, so ordinary calls or identifiers named like components remain expressions.
