# Bengali grammar

Step 7 adds Bengali as a first-class GOGO surface grammar without adding a second parser, AST, lexer, diagnostic system, formatter, or compiler pipeline.

## Architecture

Bengali source flows through the same compiler architecture as English source:

1. the existing lexer emits canonical token kinds with exact `Token.Text`;
2. the active `grammar.Vocabulary` maps Bengali surface words to canonical `grammar.Keyword` symbols;
3. the existing parser consumes those canonical keyword meanings;
4. the existing semantic AST is produced; and
5. the existing compiler session, diagnostics, source-location, and tooling paths remain in use.

Keyword recognition is data-driven by the Bengali vocabulary map. Unknown Bengali words remain identifiers. There is no source-file translation pass and no global active-language state; each parser or compiler session owns its selected vocabulary.

## Vocabulary

| Canonical keyword | Bengali forms |
| --- | --- |
| `KeywordCreate` | `তৈরি`, `ঘোষণা` |
| `KeywordVariable` | `চলক`, `ধরি` |
| `KeywordConstant` | `ধ্রুবক`, `অপরিবর্তনীয়` |
| `KeywordFunction` | `ফাংশন`, `কাজ` |
| `KeywordAs` | `হিসেবে`, `রূপে` |
| `KeywordReturn` | `ফেরত`, `ফেরাও` |
| `KeywordIf` | `যদি` |
| `KeywordElse` | `নইলে`, `অন্যথায়` |
| `KeywordImport` | `আমদানি`, `ইমপোর্ট` |
| `KeywordFrom` | `থেকে` |
| `KeywordComponent` | `কম্পোনেন্ট`, `উপাদান` |

Aliases map to the same canonical keywords as their primary forms. They do not create separate semantic AST nodes.

## Supported forms

Declarations use the same semantic forms as English:

```gogo
তৈরি চলক ব্যবহারকারী হিসেবে "Alex"
ধরি বার্তা রূপে Text হিসেবে ব্যবহারকারী
তৈরি ধ্রুবক সংখ্যা হিসেবে 42
```

Functions, parameters, return types, returns, calls, and blocks use existing punctuation and the Bengali vocabulary:

```gogo
তৈরি ফাংশন শুভেচ্ছা(নাম হিসেবে Text) হিসেবে Text {
  ফেরত নাম
}

শুভেচ্ছা("Alex")
```

Control flow uses Bengali `if` and `else` forms:

```gogo
যদি নাম {
  ফেরত নাম
} নইলে {
  ফেরাও fallback
}
```

Components use the existing component AST and property grammar:

```gogo
তৈরি কম্পোনেন্ট কার্ড(শিরোনাম হিসেবে নাম) {
  শুভেচ্ছা(শিরোনাম)
}
```

Imports localize grammar keywords only. Module paths remain string literals and aliases remain source identifiers:

```gogo
আমদানি "ui/card" হিসেবে ui
```

## Identifiers and Unicode

Bengali identifiers such as `ব্যবহারকারী`, `নাম`, `বার্তা`, `সংখ্যা`, and `ফলাফল` are preserved exactly as source text. Mixed identifiers with Bengali, ASCII, and legal digits are identifiers unless they exactly match a reserved word in the active vocabulary. GOGO does not normalize, transliterate, or rewrite Bengali identifiers.

Bengali keywords are reserved only while the Bengali vocabulary is active. In an English session, `তৈরি` can be an identifier. In a Bengali session, `তৈরি` resolves to `KeywordCreate`.

## Source positions and diagnostics

Bengali source uses the existing UTF-8 source model: byte offsets are zero-based, lines and columns are one-based, and spans are half-open. Unicode text before or after an error does not change diagnostic identity or source file paths.

Diagnostic codes remain language-independent, such as `G2004`. Bengali rendering uses the existing diagnostics catalog and keeps technical tokens such as `UTF-8` unchanged.

## Formatter behavior

Step 7 does not introduce localized keyword rewriting. Formatter-compatible Bengali source must preserve Bengali identifiers, keywords, string contents, indentation, line breaks, and UTF-8 text. If later formatter steps add canonical keyword output, they should do so through the active vocabulary rather than by hard-coded Bengali parser branches.

## Mixed-language projects

Multiple compiler sessions can parse different vocabularies at the same time. Explicitly mixed-language projects are supported by selecting the intended vocabulary per session or parser invocation. There is no global mutable language flag.

## Ambiguity and extension model

Bengali aliases are added only when they map unambiguously to existing canonical grammar concepts. New Bengali words should be added as `grammar.Entry` values in the Bengali vocabulary and should map to existing `grammar.Keyword` symbols unless the language itself gains a new semantic construct in a future roadmap step.
