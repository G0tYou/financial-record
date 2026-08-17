package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseIndonesianNumber parses Indonesian number format (e.g., "3.000.000" -> 3000000)
func ParseIndonesianNumber(input string) (float64, error) {
	// Remove dots (thousand separators in Indonesian format)
	cleaned := strings.ReplaceAll(input, ".", "")

	// Replace comma with dot for decimal separator (if any)
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Parse the cleaned string
	result, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse number '%s': %w", input, err)
	}

	return result, nil
}

// FormatIndonesianNumber formats a number to Indonesian format (e.g., 3000000 -> "3.000.000")
func FormatIndonesianNumber(number float64) string {
	// Convert to string with proper decimal handling
	str := fmt.Sprintf("%.2f", number)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// Add thousand separators
	var result string
	for i, c := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	// Add decimal part if exists
	if decimalPart != "" && decimalPart != "00" {
		result += "," + decimalPart
	}

	return result
}
