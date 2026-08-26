package parser

import "testing"

func TestStep14ParserRecoversMalformedEnumAndInterfaceDeclarations(t *testing.T) {
	cases := []string{
		`create enum { A } create variable x as 1`,
		`create enum State { , Ready } create variable x as 1`,
		`create enum State { Loading as } create variable x as 1`,
		`create enum State { Ready`,
		`create interface { name as String } create variable x as 1`,
		`create interface User extends { name as String } create variable x as 1`,
		`create interface User { readonly as String } create variable x as 1`,
		`create interface User { name } create variable x as 1`,
		`create interface User { name as String`,
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			p, file := parse(input)
			if len(p.Diagnostics()) == 0 {
				t.Fatal("malformed declaration recovered silently")
			}
			if file.Span.Start.Line == 0 {
				t.Fatal("invalid file span")
			}
		})
	}
}
