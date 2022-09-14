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
			verb, rem := ExtractPrintfVerb(tt.input)
			if verb != tt.expectedVerb || rem != tt.expectedRem {
				t.Errorf(`tPrintfVerb("c Remainder") == %s, %s. Expected == %s, %s.`, verb, rem, tt.expectedVerb, tt.expectedRem)
			}
		})
	}
}
