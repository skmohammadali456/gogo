# GOGO Parser and AST

Step 3 introduces the syntax tree layer between lexical tokens and semantic analysis.

## Pipeline

GOGO source is converted to tokens by the lexer, then parsed into a structured AST. The parser does not perform type checking, name resolution, or code generation.

```text
source -> lexer -> tokens -> parser -> AST -> semantic analysis
```

## AST principles

Every AST node carries a source span. This lets later compiler diagnostics point back to the original GOGO source.

The initial AST includes files, identifiers, literals, unary expressions, binary expressions, calls, blocks, expression statements, variable declarations, and return statements.

## Grammar separation

The parser accepts structural forms while the language grammar layer will eventually map English, Bengali, and Hindi surface forms to the same semantic constructs. Keyword recognition is therefore not added to the lexer.

## Expression parsing

Expressions use precedence climbing. Arithmetic, comparison, equality, logical, bitwise, shift, and nullish operators are represented as binary expressions. Prefix operators become unary expressions. Function calls contain an ordered argument list.

## Error recovery

Parser errors use structured diagnostics with stable G200x codes, human-readable messages, hints, and source spans. Statement recovery skips to a semicolon, closing brace, or end of file so one malformed construct does not necessarily terminate the entire parse.

## Scope of Step 3

This step establishes the parser and AST foundation. Full language grammar, type checking, name resolution, generics, pattern matching, modules, and UI syntax are implemented by later roadmap steps.
