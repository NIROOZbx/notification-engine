package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var camelCaseRegex = regexp.MustCompile(`([a-z0-9])([A-Z])`)


func FormatEventType(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	input = camelCaseRegex.ReplaceAllString(input, "${1}_${2}")

	input = nonAlphanumericRegex.ReplaceAllString(input, "_")

	input = strings.ToLower(input)

	for strings.Contains(input, "__") {
		input = strings.ReplaceAll(input, "__", "_")
	}

	return strings.Trim(input, "_")
}