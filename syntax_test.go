package main

import "testing"

func TestLoad(t *testing.T) {
	syntaxes := LoadSyntaxesFromFile("syntax.json")
	if len(syntaxes) == 0 {
		t.Errorf("Failed to load syntaxes from 'syntax.json' or no syntaxes in file.")
	}
}
