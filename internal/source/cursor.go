package source

import "unicode/utf8"

// Cursor walks a UTF-8 source string while preserving byte offsets.
type Cursor struct {
	Text   string
	Offset int
}

func NewCursor(text string) *Cursor {
	return &Cursor{Text: text}
}

func (c *Cursor) Done() bool {
	return c.Offset >= len(c.Text)
}

func (c *Cursor) Peek() (rune, int) {
	if c.Done() {
		return 0, 0
	}
	r, size := utf8.DecodeRuneInString(c.Text[c.Offset:])
	return r, size
}

func (c *Cursor) Advance() (rune, int) {
	r, size := c.Peek()
	if size > 0 {
		c.Offset += size
	}
	return r, size
}

func (c *Cursor) Match(text string) bool {
	if len(text) == 0 || c.Offset+len(text) > len(c.Text) {
		return false
	}
	if c.Text[c.Offset:c.Offset+len(text)] != text {
		return false
	}
	c.Offset += len(text)
	return true
}

func (c *Cursor) Position() Position {
	return PositionAt(c.Text, c.Offset)
}

func (c *Cursor) Slice(start int) string {
	if start < 0 {
		start = 0
	}
	if start > c.Offset {
		start = c.Offset
	}
	return c.Text[start:c.Offset]
}
