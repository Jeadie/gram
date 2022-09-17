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
		{
			input:    "%Red.s normal text%Red%hardcoded string%",
			output:   "\u001B[91mplaceholder\u001B[0m normal text\u001B[91mhardcoded string\u001B[0m",
			fmtArgs:  []any{"placeholder"},
			testName: "Colour gets reset between colour calls",
		},
		{
			input:    "%DarkBlue%STATUS BAR --% (%Blue.d, %Blue.d)",
			output:   "\u001B[34mSTATUS BAR --\u001B[0m (\u001B[94m1\u001B[0m, \u001B[94m2\u001B[0m)",
			fmtArgs:  []any{1, 2},
			testName: "Test start of status bar",
		},
		{
			input:    "%DarkBlue%STATUS BAR --% (%Blue.d, %Blue.d) of (%Magenta.d, %Magenta.d) %v. Row: %d. History: %d",
			output:   "\u001B[34mSTATUS BAR --\u001B[0m (\u001B[94m1\u001B[0m, \u001B[94m2\u001B[0m) of (\u001B[95m0\u001B[0m, \u001B[95m1\u001B[0m) [1 20 40]. Row: 20. History: 5",
			fmtArgs:  []any{1, 2, 0, 1, []byte{1, 20, 40}, 20, 5},
			testName: "Test status bar",
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

func TestExtractPrintfVerb(t *testing.T) {

	type testParam struct {
		input, expectedVerb, expectedRem, description string
	}
	tests := []testParam{{
		input:        "c Remainder",
		expectedVerb: "%c",
		expectedRem:  " Remainder",
		description:  "Standard Case",
	}, {
		input:        "c",
		expectedVerb: "%c",
		expectedRem:  "",
		description:  "No remainder",
	}, {
		input:        "",
		expectedVerb: "",
		expectedRem:  "",
		description:  "No remainder or verb",
	}}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			verb, rem := extractPrintfVerb(tt.input)
			if verb != tt.expectedVerb || rem != tt.expectedRem {
				t.Errorf(`tPrintfVerb("c Remainder") == %s, %s. Expected == %s, %s.`, verb, rem, tt.expectedVerb, tt.expectedRem)
			}
		})
	}
}

func TestAllIndices(t *testing.T) {
	type testParam struct {
		s, sub, description string
		expectedOutput      []int
	}
	tests := []testParam{{
		s:              "func TestAllIndices(t *testing.T) {",
		sub:            "function",
		description:    "substring no present",
		expectedOutput: []int{},
	}, {
		s:              " func TestAllIndices(t *testing.T) {",
		sub:            "func",
		description:    "Find initial string",
		expectedOutput: []int{1},
	}, {
		s:              "func TestAllIndices(t *testing.T) {",
		sub:            "TestAllIndices",
		description:    "Find substring within",
		expectedOutput: []int{5},
	}, {
		s:              "func AllWordIndices(s string, sub string) []int {",
		sub:            "string",
		description:    "Find multiple indices",
		expectedOutput: []int{18, 30},
	}, {
		s:              "func AllWordIndices(s stringstring) []int {",
		sub:            "string",
		description:    "Find multiple indices",
		expectedOutput: []int{18, 24},
	}}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output := AllWordIndices(tt.s, tt.sub)
			if len(output) != len(tt.expectedOutput) {
				t.Errorf("Expected %d. Received %d", tt.expectedOutput, output)
			}

			for i := range tt.expectedOutput {
				if output[i] != tt.expectedOutput[i] {
					t.Errorf("Expected %d. Received %d", tt.expectedOutput, output)
				}
			}
		})
	}
}

func TestApplyColours(t *testing.T) {
	type testParam struct {
		s, expectedOutput, description string
		highlight                      []Colour
	}

	tests := []testParam{{
		s:              "This is a string.",
		highlight:      make([]Colour, len("This is a string.")),
		expectedOutput: "This is a string.",
		description:    "No highlights",
	}}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output := ApplyColours(tt.s, tt.highlight)
			if output != tt.expectedOutput {
				t.Errorf("Expected %s. Received %s.", tt.expectedOutput, output)
			}
		})
	}
}
