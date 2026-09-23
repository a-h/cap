package table

import (
	"testing"
)

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "a lowercase word is capitalized", input: "capability", expected: "Capability"},
		{name: "an already-capitalized word is unchanged", input: "Requirement", expected: "Requirement"},
		{name: "an empty string is returned unchanged", input: "", expected: ""},
		{name: "a hyphenated kind capitalizes only the first letter", input: "external-system", expected: "External-system"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
}
