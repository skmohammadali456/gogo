# GOGO Compiler Source Model

The compiler represents source locations using byte offsets plus one-based line and column values.

## File identity

A `source.FileMap` assigns a stable numeric ID to every path in a compilation session. Re-adding the same path updates its text while retaining its ID.

## Positions and spans

`source.Position` contains:

- `Offset`: zero-based byte offset.
- `Line`: one-based line number.
- `Column`: one-based UTF-8 rune column.

`source.Span` is a half-open range `[Start, End)` and is attached to compiler diagnostics and future AST nodes.

## UTF-8 cursor

`source.Cursor` walks source text by Unicode code point while retaining byte offsets. This is important for Bengali, Hindi, and other Unicode source text. The lexer will use this cursor rather than indexing source as ASCII bytes.

## Design rule

Compiler phases must pass source spans forward. Diagnostics must point back to the original GOGO source rather than generated target code.
