package main

import (
	"strings"
)

type Row struct {
	src []byte
}

func ConstructRow(s string) Row {
	return Row{
		src: []byte(s),
	}
}

func (r Row) Render() string {
	srcStr := string(r.src)
	return strings.ReplaceAll(srcStr, "\t", "    ")
}

func (r *Row) SetCharAt(renderI uint, b byte) {

	// Convert render index to src index.
	j := 0
	for i := uint(0); i < renderI; i++ {
		j++
		if r.src[i] == '\t' {
			i += 4
		} else {
			i++
		}
	}
	r.src[j] = b
}

// RenderLen returns the length of the rendered row.
func (r Row) RenderLen() uint {
	return uint(len(r.Render()))
}
