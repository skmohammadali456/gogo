package diagnostics

import "github.com/skmohammadali786/gogo/internal/source"

type Severity uint8

const (
	Info Severity = iota
	Warning
	Error
)

type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	Hint     string
	Span     source.Span
}

type Bag struct {
	items []Diagnostic
}

func (b *Bag) Add(d Diagnostic) {
	b.items = append(b.items, d)
}

func (b *Bag) HasErrors() bool {
	for _, d := range b.items {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

func (b *Bag) All() []Diagnostic {
	out := make([]Diagnostic, len(b.items))
	copy(out, b.items)
	return out
}
