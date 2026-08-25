package diagnostics

type Code string

const (
	CodeInvalidSourcePath Code = "G0001"
	CodeInvalidSourceFile Code = "G0002"
	CodeLexMissingFile    Code = "G0003"
	CodeParseMissingFile  Code = "G0004"
	CodeReadFile          Code = "G0005"
	CodeInvalidCharacter  Code = "G1000"
	CodeUnclosedComment   Code = "G1001"
	CodeInvalidNumber     Code = "G1002"
	CodeInvalidString     Code = "G1004"
	CodeInvalidToken      Code = "G2000"
)

type CatalogEntry struct {
	Code     Code
	Severity Severity
	Messages map[Locale]string
	Hints    map[Locale]string
}

var Catalog = map[Code]CatalogEntry{
	"G0001": entry("I need a source file path before I can compile this file.", "এই ফাইল কম্পাইল করার আগে একটি সোর্স ফাইল পথ দরকার।", "इस फ़ाइल को कंपाइल करने से पहले स्रोत फ़ाइल पथ चाहिए।", "Give the source file a non-empty path, such as main.gogo.", "সোর্স ফাইলকে main.gogo-এর মতো একটি খালি নয় এমন পথ দিন।", "स्रोत फ़ाइल को main.gogo जैसा खाली न होने वाला पथ दें।"),
	"G0002": entry("This source file is not valid UTF-8 or has invalid source metadata.", "এই সোর্স ফাইলটি বৈধ UTF-8 নয় অথবা এর মেটাডেটা অবৈধ।", "यह स्रोत फ़ाइल मान्य UTF-8 नहीं है या इसका मेटाडेटा अमान्य है।", "Save the file as UTF-8 and make sure its path is valid.", "ফাইলটি UTF-8 হিসেবে সংরক্ষণ করুন এবং পথটি বৈধ কিনা নিশ্চিত করুন।", "फ़ाइल को UTF-8 के रूप में सहेजें और पथ मान्य रखें।"),
	"G0003": entry("The compiler cannot lex a source file that is not in this session.", "এই সেশনে নেই এমন সোর্স ফাইল লেক্স করা যাবে না।", "इस सत्र में न होने वाली स्रोत फ़ाइल को लेक्स नहीं किया जा सकता।", "Add the file to the compilation session before lexing it.", "লেক্স করার আগে ফাইলটি কম্পাইলেশন সেশনে যোগ করুন।", "लेक्स करने से पहले फ़ाइल को कंपाइलेशन सत्र में जोड़ें।"),
	"G0004": entry("The compiler cannot parse a source file that is not in this session.", "এই সেশনে নেই এমন সোর্স ফাইল পার্স করা যাবে না।", "इस सत्र में न होने वाली स्रोत फ़ाइल को पार्स नहीं किया जा सकता।", "Add the file to the compilation session before parsing it.", "পার্স করার আগে ফাইলটি কম্পাইলেশন সেশনে যোগ করুন।", "पार्स करने से पहले फ़ाइल को कंपाइलेशन सत्र में जोड़ें।"),
	"G0005": entry("I could not read this source file.", "আমি এই সোর্স ফাইলটি পড়তে পারিনি।", "मैं यह स्रोत फ़ाइल नहीं पढ़ सका।", "Check the file path and permissions.", "ফাইলের পথ ও অনুমতি পরীক্ষা করুন।", "फ़ाइल पथ और अनुमतियाँ जाँचें।"),
	"G1000": entry("I don't recognize this character yet.", "আমি এই অক্ষরটি এখনও চিনতে পারছি না।", "मैं इस अक्षर को अभी नहीं पहचानता।", "Check the spelling or use a supported GOGO operator or punctuation mark.", "বানান পরীক্ষা করুন অথবা সমর্থিত GOGO অপারেটর বা বিরামচিহ্ন ব্যবহার করুন।", "वर्तनी जाँचें या समर्थित GOGO ऑपरेटर या विराम चिह्न इस्तेमाल करें।"),
	"G1001": entry("This block comment never closes.", "এই ব্লক মন্তব্যটি শেষ হয়নি।", "यह ब्लॉक टिप्पणी बंद नहीं हुई है।", "Add */ to close the comment.", "মন্তব্য বন্ধ করতে */ যোগ করুন।", "टिप्पणी बंद करने के लिए */ जोड़ें।"),
	"G1002": entry("This number is not valid.", "এই সংখ্যাটি বৈধ নয়।", "यह संख्या मान्य नहीं है।", "Use a valid GOGO numeric literal.", "একটি বৈধ GOGO সংখ্যার লিটারাল ব্যবহার করুন।", "मान्य GOGO संख्यात्मक लिटरल इस्तेमाल करें।"),
	"G1004": entry("This string is not valid.", "এই স্ট্রিংটি বৈধ নয়।", "यह स्ट्रिंग मान्य नहीं है।", "Use a valid quoted string literal.", "একটি বৈধ উদ্ধৃত স্ট্রিং লিটারাল ব্যবহার করুন।", "मान्य उद्धृत स्ट्रिंग लिटरल इस्तेमाल करें।"),
	"G2000": entry("I cannot parse this invalid token.", "আমি এই অবৈধ টোকেনটি পার্স করতে পারছি না।", "मैं इस अमान्य टोकन को पार्स नहीं कर सकता।", "Fix the earlier lexical error and try again.", "আগের লেক্সিক্যাল ত্রুটিটি ঠিক করে আবার চেষ্টা করুন।", "पहली lexical त्रुटि ठीक करें और फिर प्रयास करें।"),
	"G2001": entry("I expected 'variable' or 'function' after 'create'.", "'create' এর পরে 'variable' অথবা 'function' প্রত্যাশিত।", "'create' के बाद 'variable' या 'function' अपेक्षित है।", "Write create variable name as value, or create function name(...).", "create variable name as value অথবা create function name(...) লিখুন।", "create variable name as value या create function name(...) लिखें।"),
	"G2002": entry("I expected a variable name.", "আমি একটি ভেরিয়েবলের নাম প্রত্যাশা করেছি।", "मुझे variable का नाम अपेक्षित था।", "Give the variable a name.", "ভেরিয়েবলটির একটি নাম দিন।", "variable को नाम दें।"),
	"G2003": entry("I expected 'as' after the variable name.", "ভেরিয়েবল নামের পরে 'as' প্রত্যাশিত।", "variable नाम के बाद 'as' अपेक्षित है।", "Write create variable name as value.", "create variable name as value লিখুন।", "create variable name as value लिखें।"),
	"G2004": entry("I expected a value after as.", "'as' এর পরে একটি মান প্রত্যাশিত।", "'as' के बाद एक मान अपेक्षित है।", "Give the variable an initial value.", "ভেরিয়েবলটির প্রাথমিক মান দিন।", "variable को प्रारंभिक मान दें।"),
	"G2005": entry("I expected a value after 'return'.", "'return' এর পরে একটি মান প্রত্যাশিত।", "'return' के बाद एक मान अपेक्षित है।", "Return a value or use a grammar form that explicitly permits an empty return.", "একটি মান return করুন অথবা খালি return অনুমতি দেয় এমন ব্যাকরণ ব্যবহার করুন।", "एक मान return करें या ऐसी grammar form इस्तेमाल करें जो खाली return की अनुमति देती है।"),
	"G2006": entry("This block is missing its closing brace.", "এই ব্লকের বন্ধনী } নেই।", "इस block का closing brace नहीं है।", "Add } to close the block.", "ব্লক বন্ধ করতে } যোগ করুন।", "block बंद करने के लिए } जोड़ें।"),
	"G2007": entry("I expected an expression after this operator.", "এই অপারেটরের পরে একটি expression প্রত্যাশিত।", "इस operator के बाद expression अपेक्षित है।", "Add a value after the operator.", "অপারেটরের পরে একটি মান যোগ করুন।", "operator के बाद एक मान जोड़ें।"),
	"G2008": entry("I expected an expression after this unary operator.", "এই unary অপারেটরের পরে একটি expression প্রত্যাশিত।", "इस unary operator के बाद expression अपेक्षित है।", "Add a value after the operator.", "অপারেটরের পরে একটি মান যোগ করুন।", "operator के बाद एक मान जोड़ें।"),
	"G2009": entry("I expected an expression inside these parentheses.", "এই parentheses-এর ভিতরে একটি expression প্রত্যাশিত।", "इन parentheses के भीतर expression अपेक्षित है।", "Add an expression between ( and ).", "( এবং ) এর মধ্যে একটি expression যোগ করুন।", "( और ) के बीच expression जोड़ें।"),
	"G2010": entry("This expression or call is missing a closing parenthesis.", "এই expression অথবা call-এর closing parenthesis নেই।", "इस expression या call में closing parenthesis नहीं है।", "Add ) to close it.", "এটি বন্ধ করতে ) যোগ করুন।", "इसे बंद करने के लिए ) जोड़ें।"),
	"G2011": entry("I expected a function name.", "আমি একটি function নাম প্রত্যাশা করেছি।", "मुझे function का नाम अपेक्षित था।", "Give the function a name.", "function-টির একটি নাম দিন।", "function को नाम दें।"),
	"G2012": entry("I expected '(' after the function name.", "function নামের পরে '(' প্রত্যাশিত।", "function नाम के बाद '(' अपेक्षित है।", "Add the function parameter list.", "function parameter তালিকা যোগ করুন।", "function parameter सूची जोड़ें।"),
	"G2013": entry("I expected a parameter name.", "আমি একটি parameter নাম প্রত্যাশা করেছি।", "मुझे parameter का नाम अपेक्षित था।", "Give each parameter a name.", "প্রতিটি parameter-এর একটি নাম দিন।", "हर parameter को नाम दें।"),
	"G2014": entry("I expected ')' after the function parameters.", "function parameter-এর পরে ')' প্রত্যাশিত।", "function parameters के बाद ')' अपेक्षित है।", "Close the parameter list.", "parameter তালিকা বন্ধ করুন।", "parameter सूची बंद करें।"),
	"G2015": entry("I expected a function body.", "আমি একটি function body প্রত্যাশা করেছি।", "मुझे function body अपेक्षित थी।", "Add { ... } after the function declaration.", "function declaration-এর পরে { ... } যোগ করুন।", "function declaration के बाद { ... } जोड़ें।"),
	"G2016": entry("I expected a condition after 'if'.", "'if' এর পরে একটি condition প্রত্যাশিত।", "'if' के बाद condition अपेक्षित है।", "Add an expression describing when the block should run.", "ব্লক কখন চলবে তা বোঝায় এমন expression যোগ করুন।", "block कब चले यह बताने वाला expression जोड़ें।"),
	"G2017": entry("I expected a block after the if condition.", "if condition-এর পরে একটি block প্রত্যাশিত।", "if condition के बाद block अपेक्षित है।", "Add { ... } for the conditional body.", "conditional body-এর জন্য { ... } যোগ করুন।", "conditional body के लिए { ... } जोड़ें।"),
	"G2018": entry("I expected a block after 'else'.", "'else' এর পরে একটি block প্রত্যাশিত।", "'else' के बाद block अपेक्षित है।", "Add { ... } for the else body.", "else body-এর জন্য { ... } যোগ করুন।", "else body के लिए { ... } जोड़ें।"),
	"G2020": entry("I expected a member name after property access.", "property access-এর পরে একটি member নাম প্রত্যাশিত।", "property access के बाद member नाम अपेक्षित है।", "Add a property or method name.", "property অথবা method নাম যোগ করুন।", "property या method नाम जोड़ें।"),
	"G2021": entry("I expected an index expression.", "আমি একটি index expression প্রত্যাশা করেছি।", "मुझे index expression अपेक्षित था।", "Add a value between [ and ].", "[ এবং ] এর মধ্যে একটি মান যোগ করুন।", "[ और ] के बीच एक मान जोड़ें।"),
	"G2022": entry("This index expression is missing a closing bracket.", "এই index expression-এর closing bracket নেই।", "इस index expression में closing bracket नहीं है।", "Add ] to close the index.", "index বন্ধ করতে ] যোগ করুন।", "index बंद करने के लिए ] जोड़ें।"),
	"G2023": entry("This array is missing a closing bracket.", "এই array-এর closing bracket নেই।", "इस array में closing bracket नहीं है।", "Add ] to close the array.", "array বন্ধ করতে ] যোগ করুন।", "array बंद करने के लिए ] जोड़ें।"),
	"G2024": entry("I expected an object property name.", "আমি একটি object property নাম প্রত্যাশা করেছি।", "मुझे object property नाम अपेक्षित था।", "Use an identifier or string as the property name.", "property নাম হিসেবে identifier অথবা string ব্যবহার করুন।", "property नाम के रूप में identifier या string इस्तेमाल करें।"),
	"G2025": entry("I expected ':' after the object property name.", "object property নামের পরে ':' প্রত্যাশিত।", "object property नाम के बाद ':' अपेक्षित है।", "Write key: value.", "key: value লিখুন।", "key: value लिखें।"),
	"G2026": entry("I expected an object property value.", "আমি object property-এর একটি মান প্রত্যাশা করেছি।", "मुझे object property का मान अपेक्षित था।", "Give the property a value.", "property-কে একটি মান দিন।", "property को मान दें।"),
	"G2027": entry("This object is missing a closing brace.", "এই object-এর closing brace নেই।", "इस object में closing brace नहीं है।", "Add } to close the object.", "object বন্ধ করতে } যোগ করুন।", "object बंद करने के लिए } जोड़ें।"),
	"G2028": entry("The left side of an assignment must be a variable or writable property.", "assignment-এর বাম পাশটি variable অথবা writable property হতে হবে।", "assignment का बायाँ भाग variable या writable property होना चाहिए।", "Assign to an identifier, object member, or indexed element.", "identifier, object member, অথবা indexed element-এ assign করুন।", "identifier, object member, या indexed element को assign करें।"),
	"G2029": entry("I expected an expression after '?'.", "'?' এর পরে একটি expression প্রত্যাশিত।", "'?' के बाद expression अपेक्षित है।", "Add the value to use when the condition is true.", "condition সত্য হলে ব্যবহারের মান যোগ করুন।", "condition true होने पर इस्तेमाल होने वाला मान जोड़ें।"),
	"G2030": entry("I expected ':' in this conditional expression.", "এই conditional expression-এ ':' প্রত্যাশিত।", "इस conditional expression में ':' अपेक्षित है।", "Write condition ? whenTrue : whenFalse.", "condition ? whenTrue : whenFalse লিখুন।", "condition ? whenTrue : whenFalse लिखें।"),
	"G2031": entry("I expected an expression after ':'.", "':' এর পরে একটি expression প্রত্যাশিত।", "':' के बाद expression अपेक्षित है।", "Add the value to use when the condition is false.", "condition মিথ্যা হলে ব্যবহারের মান যোগ করুন।", "condition false होने पर इस्तेमाल होने वाला मान जोड़ें।"),
	"G2032": entry("I expected a function argument.", "আমি একটি function argument প্রত্যাশা করেছি।", "मुझे function argument अपेक्षित था।", "Add an expression or close the argument list.", "একটি expression যোগ করুন অথবা argument list বন্ধ করুন।", "expression जोड़ें या argument list बंद करें।"),
	"G2033": entry("I expected an array element.", "আমি একটি array element প্রত্যাশা করেছি।", "मुझे array element अपेक्षित था।", "Add a value or close the array.", "একটি মান যোগ করুন অথবা array বন্ধ করুন।", "एक मान जोड़ें या array बंद करें।"),
	"G2034": entry("I found a closing brace without a matching block.", "মেলানো block ছাড়া একটি closing brace পাওয়া গেছে।", "मिलते block के बिना closing brace मिला।", "Remove this } or add the block that should contain it.", "এই } সরান অথবা যে block-এ এটি থাকবে সেটি যোগ করুন।", "इस } को हटाएँ या वह block जोड़ें जिसमें यह होना चाहिए।"),
	"G2035": entry("I found extra tokens after this expression statement.", "এই expression statement-এর পরে অতিরিক্ত token পাওয়া গেছে।", "इस expression statement के बाद अतिरिक्त token मिला।", "Separate statements with semicolons or use the selected grammar vocabulary.", "statement আলাদা করতে semicolon ব্যবহার করুন অথবা নির্বাচিত grammar vocabulary ব্যবহার করুন।", "statements अलग करने के लिए semicolon इस्तेमाल करें या चुनी हुई grammar vocabulary इस्तेमाल करें।"),
	"G2036": entry("I expected a quoted module path after import.", "import-এর পরে উদ্ধৃত module path প্রত্যাশিত।", "import के बाद quoted module path अपेक्षित है।", "Write import \"module\" or import \"module\" as alias.", "import \"module\" অথবা import \"module\" as alias লিখুন।", "import \"module\" या import \"module\" as alias लिखें।"),
	"G2037": entry("I expected an import alias after as.", "as-এর পরে import alias প্রত্যাশিত।", "as के बाद import alias अपेक्षित है।", "Add an identifier for the imported module alias.", "import module alias-এর জন্য identifier যোগ করুন।", "imported module alias के लिए identifier जोड़ें।"),
	"G2038": entry("I expected a component name.", "আমি একটি component নাম প্রত্যাশা করেছি।", "मुझे component नाम अपेक्षित था।", "Give the component a name.", "component-টির একটি নাম দিন।", "component को नाम दें।"),
	"G2039": entry("I expected a component property name.", "আমি component property নাম প্রত্যাশা করেছি।", "मुझे component property नाम अपेक्षित था।", "Write properties as name as value.", "property name as value আকারে লিখুন।", "properties को name as value के रूप में लिखें।"),
	"G2040": entry("I expected as after the component property name.", "component property নামের পরে as প্রত্যাশিত।", "component property नाम के बाद as अपेक्षित है।", "Write properties as name as value.", "property name as value আকারে লিখুন।", "properties को name as value के रूप में लिखें।"),
	"G2041": entry("I expected a component property value.", "আমি component property value প্রত্যাশা করেছি।", "मुझे component property value अपेक्षित था।", "Add an expression after as.", "as-এর পরে expression যোগ করুন।", "as के बाद expression जोड़ें।"),
	"G2042": entry("This component property list is missing a closing parenthesis.", "component property list-এর closing parenthesis নেই।", "component property list में closing parenthesis नहीं है।", "Add ) after the properties.", "properties-এর পরে ) যোগ করুন।", "properties के बाद ) जोड़ें।"),
	"G2043": entry("I expected a component body.", "আমি component body প্রত্যাশা করেছি।", "मुझे component body अपेक्षित थी।", "Add { ... } after the component declaration.", "component declaration-এর পরে { ... } যোগ করুন।", "component declaration के बाद { ... } जोड़ें।"),
	"G2044": entry("I expected a type name.", "আমি একটি type নাম প্রত্যাশা করেছি।", "मुझे type name अपेक्षित था।", "Use a type name such as Text, Number, Boolean, Object, or a user-defined type.", "Text, Number, Boolean, Object অথবা user-defined type ব্যবহার করুন।", "Text, Number, Boolean, Object या user-defined type इस्तेमाल करें।"),
}

func entry(en, bn, hi, hintEN, hintBN, hintHI string) CatalogEntry {
	return CatalogEntry{Severity: Error, Messages: map[Locale]string{English: en, Bengali: bn, Hindi: hi}, Hints: map[Locale]string{English: hintEN, Bengali: hintBN, Hindi: hintHI}}
}

func lookupText(code, english string, l Locale, hints bool) string {
	entry, ok := Catalog[Code(code)]
	if !ok {
		return english
	}
	m := entry.Messages
	if hints {
		m = entry.Hints
	}
	if text := m[l]; text != "" {
		return text
	}
	if text := m[English]; text != "" {
		return text
	}
	return english
}
