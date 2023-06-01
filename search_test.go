package main

import (
	"testing"
)

func TestSearchRows(t *testing.T) {
	rows := []Row{
		{src: []byte("Hello, world!")},
		{src: []byte("Go is awesome")},
		{src: []byte("Hello, Go")},
		{src: []byte("Goodbye, world!")},
	}
	q := "Go"

	results := SearchRows(rows, q)

	// Collect all results.
	var res []SearchResult
	for r := range results {
		res = append(res, r)
	}

	// Expecting 3 results.
	if len(res) != 3 {
		t.Errorf("Expected 3 results, got %d", len(res))
	}

	// Check the results.
	if string(res[0].rowRef.src) != "Go is awesome" || res[0].startI != 0 || res[0].rowI != 1 {
		t.Errorf("Unexpected result: %+v", res[0])
	}
	if string(res[1].rowRef.src) != "Hello, Go" || res[1].startI != 7 || res[1].rowI != 2 {
		t.Errorf("Unexpected result: %+v", res[1])
	}
	if string(res[2].rowRef.src) != "Goodbye, world!" || res[2].startI != 0 || res[2].rowI != 3 {
		t.Errorf("Unexpected result: %+v", res[1])
	}
}
