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

func (r Row) Export() []byte {
	return append(r.src, '\n')
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
	return 0
}

func (r *Row) AddCharAt(renderI uint, b byte) {
	j := r.getSrcIndex(renderI)

	x := r.src[:j]
	y := r.src[j:]
	z := strings.Join([]string{string(x), string(y)}, string(b))
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
