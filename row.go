package main

import "strings"

type Row struct {
	render string
	src    string
}

func ConstructRow(s string) Row {
	render := strings.ReplaceAll(s, "\t", "    ")

	return Row{
		render: render,
		src:    s,
	}
}

func (r Row) Render() string {
	return r.render
}

// RenderLen returns the length of the rendered row.
func (r Row) RenderLen() uint {
	return uint(len(r.render))
}
