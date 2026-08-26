# Language configuration and aliases

Step 9 introduces one project-level configuration file, `gogo.json`, for language-facing compiler settings. The file is data only: it never executes commands, loads plugins, or invokes shell hooks.

## Schema

All fields are optional. Omitting the file is valid and preserves existing English behavior. A configuration file must contain exactly one JSON object and cannot contain unknown fields.

```json
{
  "language": "english",
  "aliases": [
    { "surface": "make", "keyword": "create" }
  ],
  "strictness": "standard",
  "encoding": "utf-8",
  "target": "ast",
  "compatibility": "step9"
}
```

`language` accepts `en`/`english`, `bn`/`bengali`, and `hi`/`hindi`. The language determines the grammar vocabulary; users do not repeat a separate grammar field.

`aliases` add project spellings that resolve to existing canonical grammar keywords (`create`, `variable`, `constant`, `function`, `component`, `import`, `from`, `as`, `return`, `if`, `else`). Aliases do not create new AST semantics.

`strictness` accepts `standard`, `strict`, or `permissive`. Step 9 centralizes the setting for later diagnostics policy; the Step 1-8 grammar has no construct whose severity changes by mode yet.

`encoding` must be `utf-8`. GOGO source text remains UTF-8, including English, Bengali, and Hindi projects. Malformed UTF-8 continues to use the existing source and lexer diagnostics.

`target` is `ast`, the only target available before backend roadmap steps. The configuration type is central so later targets can be added without rewriting language selection.

`compatibility` is `step9`, the current language-configuration contract. Compiler version, language compatibility, and future grammar versions remain separate concepts.

## Discovery and precedence

The CLI resolves configuration deterministically:

1. explicit `-config path`, if provided;
2. otherwise the nearest `gogo.json` found by walking from the first source file toward the filesystem root;
3. otherwise built-in defaults.

Supported CLI overrides are `-grammar`, `-strictness`, and `-target`. If a CLI language override conflicts with project `language`, the compiler emits a configuration diagnostic instead of silently choosing one.

## Validation and diagnostics

Configuration is resolved once into an immutable `config.Resolved` value before the compiler session starts. Validation rejects unknown languages, strictness modes, encodings, targets, compatibility values, invalid alias identifiers, duplicate alias conflicts, aliases for unknown canonical keywords, and aliases that conflict with built-in reserved words. Diagnostics use stable `G300x` codes and include the configuration source path when available.

## Compiler integration

`compiler.Session` consumes `config.Resolved` and passes the selected `grammar.Vocabulary` to the shared parser. There is no global active language. Multiple sessions can compile English, Bengali, and Hindi projects concurrently without changing one another's grammar.

## Multilingual behavior

Configuration selects the active vocabulary; it does not translate source. A Bengali project accepts Bengali keywords and legal Bengali or ASCII identifiers according to the existing lexer and parser rules. English or Hindi keywords are not recognized in that project unless they are configured as explicit, validated aliases.

## Migration

Existing Step 6 English projects need no file. Add `gogo.json` only when a project wants Bengali or Hindi grammar, project aliases, or explicit documentation of strictness, target, encoding, and compatibility settings.
