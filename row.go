package main

import (
	"strings"
)

type Row struct {
	src []byte
}

func ConstructRow(s string) Row {
	return Row{
		src: []byte(strings.ReplaceAll(s, "\t", "    ")),
	}
}

// SplitAt a given rendered index into a row. Creates two new rows, original unchanged.
func (r Row) SplitAt(i uint) (*Row, *Row) {
	a := Row{src: make([]byte, i)}
	b := Row{src: make([]byte, uint(len(r.src))-i)}
	copy(a.src, r.src[:i])
	copy(b.src, r.src[i:])

	return &a, &b
}

func (r Row) Export() []byte {
	return append(r.src, '\n')
}

// RenderWithin from constraints of, starting from offset index, and being no larger than max.
func (r Row) RenderWithin(offset, max uint) string {
	l := r.Render()
	if offset > uint(len(l)) {
		return ""
	}

	if uint(len(l)) > max+offset {
		return l[offset : offset+max]
	} else {
		return l[offset:]
	}
}

func (r Row) Render() string {
	srcStr := string(r.src)
	return strings.ReplaceAll(srcStr, "\t", "    ")
}

func (r *Row) getSrcIndex(renderI uint) int {
	j := uint(0)
	for i := 0; i < len(r.src); i++ {
		if j >= renderI {
			return i
		}
		if r.src[i] == '\t' {
			j += 4
		} else {
			j++
		}
	}

	// Accomodate adding to end of line.
	if j >= renderI {
		return len(r.src)
	}
	return 0
}

func (r *Row) RemoveCharAt(renderI uint) {
	j := r.getSrcIndex(renderI)
	if j == 0 {
		// TODO: add current row to row above
		return
	} else if j-1 >= len(r.src) {

		// Delete last character in row, no characters to append
		r.src = r.src[:j-1]
	} else {
		r.src = append(r.src[:j-1], r.src[j:]...)
	}
}

func (r *Row) AddCharAt(renderI uint, b byte) {
	r.AddCharsAt(renderI, string(b))
}

func (r *Row) AddCharsAt(renderI uint, s string) {
	j := r.getSrcIndex(renderI)

	x := r.src[:j]
	y := r.src[j:]
	z := strings.Join([]string{string(x), string(y)}, s)
	r.src = []byte(z)
}

func (r *Row) GetCharAt(renderI uint) byte {
	return r.src[r.getSrcIndex(renderI)]
}

func (r *Row) SetCharAt(renderI uint, b byte) {
	j := r.getSrcIndex(renderI)
	r.src[j] = b
}

// RenderLen returns the length of the rendered row.
func (r Row) RenderLen() uint {
	return uint(len(r.Render()))
}

// GetNextWordFrom the current render index. Returns the index of the space in front of the next word. parameter nextWordRight
// Determined if next word left (false), or right (true).
func (r Row) GetNextWordFrom(renderI uint, nextWordRight bool) uint {
	if len(r.src) == 0 {
		return 0
	}

	// TODO: [BUG] Indexes src index not render index. Therefore we need to address possibility
	//   r.src[i] == '/t' and thus would shift by more than one.
	j := r.getSrcIndex(renderI)
	if nextWordRight {
		for i := j + 1; i < len(r.src); i++ {
			if r.src[i] == ' ' {
				return uint(i)
			}
		}
		return uint(len(r.src) - 1)
	} else {
		for i := j - 1; i >= 0; i-- {
			if r.src[i] == ' ' {
				return uint(i)
			}
		}
		return 0
	}
}

// RenderIndexOf string within row starting `from` . Returns -1 if not found, or render index from
// start of row (not from `from` index).
func (r *Row) RenderIndexOf(s string, from int) int {
	if from >= len(r.Render()) {
		return -1
	}

	y := strings.Index(r.Render()[from:], s)
	if y == -1 {
		return -1
	}
	return from + y
}

// Append Row ro, into Row.
func (r *Row) Append(ro *Row) {
	r.src = append(r.src, ro.src...)
}
