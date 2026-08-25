# GOGO Parser and AST

Step 3 converts lexical tokens into a structured, source-spanned abstract syntax tree. It is the structural syntax layer between lexing and later semantic analysis.

## Pipeline

```text
source -> lexer -> tokens -> parser -> AST -> semantic analysis
```

The parser owns structure, grouping, precedence, syntax diagnostics, and recovery. It does not perform type checking, name resolution, overload resolution, or code generation.

## AST principles

Every AST node carries a source span. Later compiler phases can therefore report diagnostics against the original GOGO source.

Step 3 AST coverage includes:

- files and statement lists
- identifiers and literals
- unary expressions
- binary expressions
- assignment and compound assignment expressions
- conditional expressions
- function calls and ordered arguments
- function declarations and parameters
- blocks
- expression statements
- variable declarations
- return statements
- if and else blocks
- arrays and array elements
- object literals and object properties
- required and optional member access
- index access

## Structural grammar

The current prototype supports the structural forms needed by the Step 3 parser foundation:

```text
create variable name as expression
create function name(parameter, ...) { ... }
return expression
if expression { ... } else { ... }

expression
expression = expression
expression += expression
expression ? expression : expression
expression.member
expression?.member
expression[index]
expression(arguments)
[expression, ...]
{ key: expression, ... }
```

The parser permits trailing commas in function parameter lists, calls, arrays, and objects.

## Expression parsing

Expressions use precedence climbing. Assignment operators are right associative. Exponentiation is right associative. Arithmetic, comparison, equality, logical, bitwise, shift, and nullish operators are represented as binary expressions. Prefix operators become unary expressions. Calls, member access, and indexing bind as postfix operations.

`MemberExpr.Optional` records whether member access used `?.` rather than `.`. This preserves the lexical distinction for later semantic and lowering phases.

Assignment targets are checked structurally. An assignment target must be an identifier, member expression, or index expression. The parser reports G2028 for invalid assignment targets.

## Multilingual source

The parser accepts Unicode identifiers, including Bengali and Hindi. It does not hard-code localized keyword dictionaries. Later grammar work maps localized surface forms onto the same semantic constructs.

## Error recovery

Parser errors use structured diagnostics with stable G200x codes, human-readable messages, hints, and source spans. Statement recovery advances to a semicolon, closing brace, or end of file. Top-level recovery also handles unmatched closing braces explicitly. Object recovery recognizes commas and closing braces so malformed properties do not unnecessarily destroy the surrounding object.

The parser is designed to continue after syntax errors and produce the maximum useful AST for later diagnostics.

## Scope boundary

Step 3 is intentionally structural. Type checking, name resolution, localized keyword grammars, generics, modules, pattern matching, UI syntax, IR, and code generation belong to later roadmap steps.
