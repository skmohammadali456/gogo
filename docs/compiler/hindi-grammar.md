# Hindi GOGO grammar

Step 8 adds Hindi as a first-class surface vocabulary through the existing grammar abstraction. Hindi source is lexed by the shared lexer, resolved by `grammar.Vocabulary` into canonical `grammar.Keyword` values, parsed by the shared parser, and represented by the shared semantic AST.

## Vocabulary audit

| Canonical keyword | Hindi forms | Tests |
| --- | --- | --- |
| `KeywordCreate` | `बनाओ`, `निर्माण` | `TestStep8HindiVocabularyMappings`, parser declaration/component tests |
| `KeywordVariable` | `चर`, `मान` | `TestStep8HindiDeclarationsAndUnicodeIdentifiers` |
| `KeywordConstant` | `स्थिर`, `अचर` | `TestStep8HindiDeclarationsAndUnicodeIdentifiers` |
| `KeywordFunction` | `फलन`, `कार्य`, `फ़ंक्शन` | `TestStep8HindiFunctionsControlFlowCallsComponentsImports` |
| `KeywordAs` | `रूप`, `जैसा`, `के_रूप_में` | declaration, parameter, import alias, and component property tests |
| `KeywordReturn` | `लौटाओ`, `वापस` | function/control-flow tests |
| `KeywordIf` | `अगर`, `यदि` | control-flow and recovery tests |
| `KeywordElse` | `वरना`, `अन्यथा` | control-flow tests |
| `KeywordImport` | `आयात`, `लाओ` | import tests |
| `KeywordFrom` | `से` | vocabulary audit; reserved for module forms supported by the canonical grammar |
| `KeywordComponent` | `घटक`, `अवयव` | component tests |

All Hindi forms are aliases for canonical grammar keywords. They do not create Hindi-specific parser states, AST nodes, or compiler behavior.

## Supported constructs

Hindi supports the same Step 6 and Step 7 constructs currently implemented for English and Bengali:

- readable declarations, concise declarations, constants, type annotations, initializers, and assignment expressions where the shared parser permits them;
- function declarations, parameter lists, parameter types, return types, return statements, function bodies, and calls;
- `if`/`else` control flow through `अगर`/`यदि` and `वरना`/`अन्यथा`;
- imports with string-literal module paths and identifier aliases;
- component declarations, property lists, typed/value-like property expressions, component bodies, and nested statements supported by the shared AST;
- expressions, calls, member/index access, arrays, objects, conditional expressions, and operators supported by the shared parser.

## Identifier and Unicode behavior

The lexer remains language-neutral. Devanagari identifiers such as `उपयोगकर्ता`, `नाम`, `संदेश`, `परिणाम`, and `संख्या` are preserved exactly as token text and source bytes. Mixed identifiers such as `userहिंदी`, `हिंदीUser`, and `उपयोगकर्ता123` are accepted when they satisfy the shared lexical identifier rules. Identifiers are not transliterated, ASCII-normalized, or translated.

Source spans remain byte-offset based with one-based human-readable line and Unicode column positions. Hindi tests verify `PositionAt` consistency for multiline UTF-8 source and diagnostics with Hindi text before and after the error span.

## Diagnostics

Hindi uses the shared diagnostics subsystem and stable language-independent diagnostic codes. Rendering with Hindi locale (`hi`) localizes human messages and hints, while codes such as `G2004` remain unchanged. JSON diagnostics include the same code, severity, source span, file path, and language metadata used by English and Bengali.

## Mixed-language behavior and ambiguity rules

The active vocabulary controls keyword recognition. Hindi keywords are reserved only in Hindi grammar sessions. English and Bengali sessions may use Hindi words as identifiers where lexical rules allow them. Conversely, unknown Hindi words remain identifiers in Hindi sessions. Cross-language keywords are not silently interpreted unless the active vocabulary explicitly maps that surface form.

Import paths are string literals and are never translated. Aliases and all other user identifiers are preserved exactly.

## Formatter compatibility

The full formatter is assigned to a later roadmap step. Step 8 does not implement a formatter. Hindi remains formatter-compatible because Unicode source, Hindi identifiers, canonical grammar token resolution, language-independent AST semantics, source spans, and diagnostics all flow through the same architecture that the future formatter will use.

## Extension model

Future Hindi aliases should be added as entries in `internal/grammar` that map surface text to existing canonical `grammar.Keyword` values whenever semantics are unchanged. New language features should add canonical keywords and shared parser/AST/compiler support first, then vocabulary entries for each supported surface grammar.
