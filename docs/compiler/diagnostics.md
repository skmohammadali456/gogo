# Compiler diagnostics

GOGO Step 4 promotes diagnostics to a compiler subsystem shared by lexer, parser, compiler sessions, the CLI, and future editor tooling.

## Model

A diagnostic has a stable `G` code registered in the diagnostic catalog, severity (`error`, `warning`, `info`, or `hint`), an English base message, file identity, a primary source span, labels, notes, hints, suggestions, and optional fix-it edits. Spans remain half-open byte ranges backed by UTF-8-safe line and column positions from the source package.

## Rendering

Human output is deterministic and includes:

- `severity[code]: message` header.
- `path:line:column` location.
- Source snippet lines.
- Carets under the primary span, including multiline spans.
- Notes, hints, and suggestions.

JSON output serializes the same ordered and deduplicated model for editor and build-server integrations.

## Language support

Diagnostics keep English as the canonical message set for stable tests and tooling. The catalog stores localized messages for stable codes, and the renderer accepts `en`, `bn`, and `hi` locales so CLI users can request English, Bengali, or Hindi diagnostic output without changing lexer or parser internals. Full grammar localization is intentionally left to the later grammar steps.

## Ordering and deduplication

Diagnostic bags return diagnostics sorted by source start offset, end offset, and code. Duplicate diagnostics with the same code, span, severity, and message are emitted once.

## CLI

`gogo file.gogo` parses files and prints human diagnostics. `gogo -json file.gogo` prints machine-readable JSON. `gogo -locale bn file.gogo` and `gogo -locale hi file.gogo` select Bengali or Hindi rendering.

## Step 4 validation checklist

This subsystem implements the Step 4 requirements as follows:

1. Diagnostic model: `Diagnostic`, `Label`, `Suggestion`, and `FixIt`.
2. Stable diagnostic codes: the `Catalog` maps existing `G000x`, `G100x`, and `G200x` codes.
3. Severity levels: `error`, `warning`, `info`, and `hint`.
4. Source spans: every diagnostic carries a half-open `source.Span`.
5. Primary labels: span-only diagnostics normalize to a primary label.
6. Secondary labels: related label data can be attached and rendered.
7. Notes: rendered after snippets.
8. Hints: legacy and structured hints are both supported.
9. Suggestions: user-facing suggested actions are rendered and serialized.
10. Fix-it edits: suggestions can carry replacement edits.
11. Source snippets: human output includes source lines.
12. Caret rendering: human output underlines the selected span.
13. Multiline diagnostics: snippets cover every line in a multiline span.
14. UTF-8 safe source positions: columns are rune-counted while spans retain byte offsets.
15. English diagnostics: English is the canonical message set.
16. Bengali diagnostics: Bengali catalog entries are available for stable codes.
17. Hindi diagnostics: Hindi catalog entries are available for stable codes.
18. Lexer diagnostic integration: lexer diagnostics merge through compiler sessions.
19. Parser diagnostic integration: parser diagnostics merge through compiler sessions.
20. Compiler diagnostic integration: session-level diagnostics share the same bag and renderer.
21. Human-friendly CLI output: `gogo file.gogo` prints text diagnostics.
22. JSON diagnostic output: `gogo -json file.gogo` prints structured diagnostics.
23. Deterministic diagnostic ordering: bags sort by file, span, and code.
24. Diagnostic deduplication: duplicate diagnostics are emitted once.
25. Golden diagnostic tests: multiline rendering has a checked golden fixture.
26. English/Bengali/Hindi regression tests: tests exercise all three rendering locales.
27. Documentation: this document is the Step 4 diagnostic contract.
