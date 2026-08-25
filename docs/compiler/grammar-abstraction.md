# Grammar abstraction

Step 5 adds a grammar layer between lexical identifiers and parser rules. The lexer still emits canonical token kinds plus exact source text and spans. The parser receives one active `grammar.Vocabulary`, resolves identifier tokens through that vocabulary, and consumes canonical `grammar.Keyword` symbols such as `KeywordCreate`, `KeywordVariable`, `KeywordFunction`, `KeywordAs`, `KeywordReturn`, `KeywordIf`, and `KeywordElse`.

The AST remains the single semantic AST in `internal/ast`; it does not record the surface keyword that introduced a declaration or statement. English, Bengali, and Hindi sources therefore produce equivalent semantic AST structures when they use equivalent identifiers and literals. Source spans continue to point at the original user text because the grammar layer only performs keyword lookup and does not rewrite token text.

## Built-in vocabularies

The built-in vocabularies are data-driven tables in `internal/grammar`:

| Semantic keyword | English | Bengali | Hindi |
| --- | --- | --- | --- |
| create | `create` | `তৈরি` | `बनाओ` |
| variable | `variable` | `চলক` | `चर` |
| function | `function` | `ফাংশন` | `फ़ंक्शन` |
| as | `as` | `হিসেবে` | `के_रूप_में` |
| return | `return` | `ফেরত` | `लौटाओ` |
| if | `if` | `যদি` | `अगर` |
| else | `else` | `নইলে` | `वरना` |

Only words in the active vocabulary are reserved for parser keyword positions. Unknown words remain normal identifiers, preserving Unicode identifier behavior.

## Selecting a vocabulary

Compiler sessions default to English. A caller can select another vocabulary with `compiler.WithGrammarLanguage(grammar.Bengali)`, `compiler.WithGrammarLanguage(grammar.Hindi)`, or `compiler.WithGrammarVocabulary`. The CLI exposes the same choice through `-grammar en`, `-grammar bn`, or `-grammar hi`.

Vocabulary state is stored per parser/session instance. There is no global mutable language mode, so multiple sessions can parse different languages safely.

## Mixed language behavior

Mixed keyword vocabularies are not accepted in Step 5. For example, English `create` is an identifier in Bengali grammar mode rather than a Bengali keyword. If that creates invalid syntax, the normal parser diagnostics are emitted through the centralized diagnostics subsystem.

## Adding another vocabulary

Add a new `grammar.Language` value and a `NewVocabulary` entry mapping surface terms to the existing canonical `grammar.Keyword` symbols. Parser logic should not change unless the semantic language itself gains a new construct in a later GOGO step.
