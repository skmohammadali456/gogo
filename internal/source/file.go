package source

import "unicode/utf8"

// FileMap owns source files for one compiler session and gives every file a stable ID.
type FileMap struct {
	files  []File
	byPath map[string]uint32
}

func NewFileMap() *FileMap {
	return &FileMap{byPath: make(map[string]uint32)}
}

func (m *FileMap) Add(path, text string) uint32 {
	if id, ok := m.byPath[path]; ok {
		m.files[id-1].Text = text
		return id
	}
	id := uint32(len(m.files) + 1)
	m.files = append(m.files, File{ID: id, Path: path, Text: text})
	m.byPath[path] = id
	return id
}

func (m *FileMap) Get(id uint32) (File, bool) {
	if id == 0 || int(id) > len(m.files) {
		return File{}, false
	}
	return m.files[id-1], true
}

func (m *FileMap) Count() int { return len(m.files) }

// LineStartOffsets returns byte offsets for the beginning of each line.
// The first line always starts at offset zero.
func LineStartOffsets(text string) []int {
	offsets := []int{0}
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == '\n' {
			offsets = append(offsets, i+size)
		}
		i += size
	}
	return offsets
}

// PositionAt converts a byte offset into a one-based source position.
func PositionAt(text string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}

	line, lineStart := 1, 0
	for i := 0; i < offset; {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == '\n' {
			line++
			lineStart = i + size
		}
		i += size
	}

	column := 1
	for i := lineStart; i < offset; {
		_, size := utf8.DecodeRuneInString(text[i:])
		column++
		i += size
	}
	return Position{Offset: offset, Line: line, Column: column}
}
