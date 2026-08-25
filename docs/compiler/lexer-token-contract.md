# GOGO Lexer and Token Contract

Step 2 defines the lexical contract consumed by the grammar layer.

## Token model

Every token contains:

- a canonical `token.Kind`
- the exact source `Text`
- a half-open `source.Span`

The lexer never assigns language keywords. Words such as `create`, `variable`, and future Bengali or Hindi surface keywords remain `Identifier` tokens. Keyword recognition belongs to the grammar layer.

## Identifiers

Identifiers support Unicode letters, combining marks, digits after the first character, and `_`.

This allows source such as:

```gogo
create variable नाम as "दिल्ली"
create variable নাম as "বাংলা"
```

## Numbers

GOGO supports:

- integers, such as `42`
- decimals, such as `10.5`
- exponent notation, such as `1e6` and `2.5E-3`

A decimal point must be followed by a digit. An exponent must contain digits and may contain `+` or `-` after `e` or `E`.

Malformed forms produce lexical diagnostics rather than silently becoming valid numbers.

## Strings

Single and double quoted strings are supported. Escapes are scanned and validated for the supported escape set: `n`, `r`, `t`, `b`, `f`, `v`, `0`, `\\`, `"`, and `'`.

An unterminated string or unsupported escape produces a diagnostic and an `Invalid` token so lexing can continue.

## Comments

`//` starts a line comment. `/*` and `*/` delimit block comments.

Comments are skipped by default. Tooling can call `IncludeComments(true)` to preserve them as `Comment` tokens.

## Recovery and UTF-8

Invalid UTF-8 is diagnosed at the lexer boundary. Unknown characters become `Invalid` tokens and advance the cursor, preventing a single bad character from trapping the lexer.

## Compiler integration

`compiler.Session.LexFile` runs the lexer for a registered source file and copies lexer diagnostics into the session diagnostic bag. Tokens retain source spans from the original source file.

## Acceptance tests

The Step 2 test suite covers basic programs, Unicode identifiers, numbers, operators, comments, malformed numbers, invalid escapes, unterminated strings, unterminated comments, malformed UTF-8, source spans, compiler-session integration, and a fuzz target that checks the lexer does not panic on arbitrary input.
