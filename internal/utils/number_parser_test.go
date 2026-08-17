package utils

import (
	"testing"
)

func TestParseIndonesianNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"3.000.000", 3000000, false},
		{"1.500.500", 1500500, false},
		{"100.000", 100000, false},
		{"50.000", 50000, false},
		{"1.000", 1000, false},
		{"500", 500, false},
		{"3.000.000,50", 3000000.50, false},
		{"1.500.000,25", 1500000.25, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := ParseIndonesianNumber(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("For input '%s', expected %f but got %f", test.input, test.expected, result)
			}
		}
	}
}

func TestFormatIndonesianNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{3000000, "3.000.000"},
		{1500500, "1.500.500"},
		{100000, "100.000"},
		{50000, "50.000"},
		{1000, "1.000"},
		{500, "500"},
		{3000000.50, "3.000.000,50"},
		{1500000.25, "1.500.000,25"},
	}

	for _, test := range tests {
		result := FormatIndonesianNumber(test.input)
		if result != test.expected {
			t.Errorf("For input %f, expected '%s' but got '%s'", test.input, test.expected, result)
		}
	}
}