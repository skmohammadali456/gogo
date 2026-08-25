package source

// Position identifies a location in a GOGO source file.
// Offset is a zero-based byte offset. Line and Column are one-based.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span identifies a half-open source range [Start, End).
type Span struct {
	Start Position
	End   Position
}

// File identifies a source file in a compilation session.
type File struct {
	ID   uint32
	Path string
	Text string
}

func NewPosition() Position {
	return Position{Line: 1, Column: 1}
}

func (p Position) IsValid() bool {
	return p.Line >= 1 && p.Column >= 1 && p.Offset >= 0
}

func (s Span) IsValid() bool {
	return s.Start.IsValid() && s.End.IsValid() && s.End.Offset >= s.Start.Offset
}
