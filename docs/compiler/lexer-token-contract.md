# GOGO Lexer and Token Contract

Step 2 defines the lexical contract consumed by the grammar layer.

## Multilingual rule

GOGO has one Unicode-aware lexer for English, Bengali, Hindi, and future surface grammars. The lexer does not assign language keywords. Words such as `create`, `variable`, `বাংলা`, and `हिन्दी` remain `Identifier` tokens. The grammar layer later maps surface vocabulary to semantic keywords.

Examples:

```gogo
create variable user as "Alex"
বাংলায় variable user as "Alex"
हिन्दी में variable user as "Alex"
```

The exact keyword vocabulary is intentionally a Step 5 to Step 8 grammar concern, not a lexer concern.

## Token model

Every token contains:

- a canonical `token.Kind`
- the exact source `Text`
- a half-open `source.Span`

The token model includes EOF, Identifier, Number, String, Invalid, punctuation, operators, and Comment trivia. Helper methods expose trivia, EOF, and invalid-token checks.

## Identifiers

Identifiers support Unicode letters, combining marks, digits after the first character, and `_`. This covers English, Bengali, Hindi, and other Unicode scripts without separate scanner implementations.

## Numbers

GOGO's lexer recognizes:

- decimal integers, such as `42`
- decimals, such as `10.5`
- exponent notation, such as `1e6` and `2.5E-3`
- numeric separators, such as `1_000`
- hexadecimal, such as `0xff`
- binary, such as `0b1010`
- octal, such as `0o755`
- integer suffix `n`, reserved for the future numeric type system

Malformed decimal, exponent, separator, or based-number forms produce lexical diagnostics.

## Strings

Single and double quoted strings are supported. Escape validation covers simple escapes such as `\\n`, `\\r`, `\\t`, `\\b`, `\\f`, `\\v`, `\\0`, `\\\\`, `\\"`, and `\\'`, hexadecimal escapes such as `\\x41`, four-digit Unicode escapes such as `\\u0041`, and braced Unicode escapes such as `\\u{1F600}`.

An unterminated string or invalid escape produces a diagnostic and an `Invalid` token so lexing can continue safely.

## Operators

The token model is intentionally broad enough for the TypeScript and JavaScript-inspired GOGO surface. It recognizes assignment, equality, strict equality, arithmetic, increment/decrement, arrows, exponentiation, remainder, comparisons, bitwise operations, logical operations, shifts, nullish coalescing, optional chaining, ternary punctuation, and spread punctuation.

Supported multi-character operators include `==`, `===`, `!=`, `!==`, `+=`, `++`, `-=`, `--`, `->`, `=>`, `*=`, `**`, `**=`, `/=`, `%=`, `<<`, `<<=`, `>>`, `>>=`, `>>>`, `>>>=`, `&=`, `&&`, `|=`, `||`, `^=`, `??`, `?.`, and `...`.

## Punctuation

Braces, parentheses, brackets, comma, dot, colon, semicolon, and question mark have dedicated token kinds.

## Whitespace and comments

Unicode whitespace is skipped. `//` starts a line comment. `/*` and `*/` delimit block comments. Comments are skipped by default. Tooling can call `IncludeComments(true)` to preserve comments as `Comment` tokens.

## Recovery and UTF-8

Invalid UTF-8 is diagnosed at the lexer boundary. Unknown characters become `Invalid` tokens and advance the cursor, preventing a single bad character from trapping the lexer.

## Compiler integration

`compiler.Session.LexFile` runs the lexer for a registered source file and copies lexer diagnostics into the session diagnostic bag. Tokens retain source spans from the original source file.

## Acceptance tests

The Step 2 test suite covers English, Bengali, and Hindi identifiers, numbers and numeric forms, the complete operator surface, punctuation, strings and Unicode escapes, invalid literals, comments, malformed UTF-8, source spans, compiler-session integration, and a fuzz target that checks the lexer does not panic on arbitrary input.
