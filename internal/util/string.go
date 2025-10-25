package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"
)

// Truncate truncates a string to the specified length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TruncateMiddle truncates a string in the middle to fit maxLen
func TruncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen < 3 {
		return s[:maxLen]
	}

	ellipsis := "..."
	sideLen := (maxLen - len(ellipsis)) / 2
	return s[:sideLen] + ellipsis + s[len(s)-sideLen:]
}

// IsBlank checks if a string is empty or contains only whitespace
func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// DefaultIfBlank returns defaultVal if s is blank
func DefaultIfBlank(s, defaultVal string) string {
	if IsBlank(s) {
		return defaultVal
	}
	return s
}

// RandomString generates a random string of the specified length
func RandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// RandomID generates a random ID (22 chars, URL-safe)
func RandomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Contains checks if a string slice contains a string
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ContainsAny checks if a string slice contains any of the given strings
func ContainsAny(slice []string, strs ...string) bool {
	for _, str := range strs {
		if Contains(slice, str) {
			return true
		}
	}
	return false
}

// Unique removes duplicates from a string slice
func Unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// SplitLines splits a string into lines
func SplitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

// JoinNonEmpty joins non-empty strings with a separator
func JoinNonEmpty(sep string, strs ...string) string {
	var nonEmpty []string
	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, sep)
}

// Indent indents each line of a string
func Indent(s string, prefix string) string {
	lines := SplitLines(s)
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// PadLeft pads a string on the left to reach the specified length
func PadLeft(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(string(pad), length-len(s)) + s
}

// PadRight pads a string on the right to reach the specified length
func PadRight(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(pad), length-len(s))
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	// Split on non-alphanumeric characters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return ""
	}

	// First word lowercase, rest title case
	result := strings.ToLower(words[0])
	for _, word := range words[1:] {
		result += strings.Title(strings.ToLower(word))
	}

	return result
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

// ExtractQuoted extracts text between quotes
func ExtractQuoted(s string) []string {
	var results []string
	inQuote := false
	var current []rune

	for _, r := range s {
		if r == '"' {
			if inQuote {
				results = append(results, string(current))
				current = []rune{}
			}
			inQuote = !inQuote
		} else if inQuote {
			current = append(current, r)
		}
	}

	return results
}

// Sanitize removes or replaces unsafe characters for filenames
func Sanitize(s string) string {
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := s
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
