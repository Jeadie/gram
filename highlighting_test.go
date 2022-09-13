package main

import (
	"fmt"
	"testing"
)

type DoCprintfParseTestCase struct {
	input, output, testName string
	fmtArgs                 []any
}

func TestDoCprintfParse(t *testing.T) {
	tests := []DoCprintfParseTestCase{
		{
			input:    "",
			output:   "",
			fmtArgs:  []any{},
			testName: "Empty input shouldn't be modified",
		},
		{
			input:    "%Red.s",
			output:   "\u001B[91mplaceholder\u001B[0m",
			fmtArgs:  []any{"placeholder"},
			testName: "placeholders get coloured correctly",
		},
		{
			input:    "%Red%hardcoded string%",
			output:   "\u001B[91mhardcoded string\u001B[0m",
			fmtArgs:  []any{},
			testName: "Hard coded strings get coloured correctly",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			payload := doCprintfParse(tt.input)
			output := fmt.Sprintf(payload, tt.fmtArgs...)
			if output != tt.output {
				t.Errorf(fmt.Sprintf("Expected %s, received %s", tt.output, output))
			}
		})
	}
}
