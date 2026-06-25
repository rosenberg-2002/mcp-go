package datautils

import (
	"crypto/md5" //nolint:gosec // used for non-security purposes (checksums/cache keys)
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// GenerateHash produces MD5 and SHA256 hashes from the input string.
// Practical uses: file integrity checks, cache keys, content-addressable storage.
func GenerateHash(text string) string {
	md5Sum := md5.Sum([]byte(text)) //nolint:gosec
	sha256Sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf(
		"Input  : %q\n%s\nMD5    : %x\nSHA256 : %x\n\nNote: MD5 should not be used for security purposes (use SHA256 instead).",
		text, strings.Repeat("─", 50), md5Sum, sha256Sum,
	)
}

// Base64Encode encodes a string to Base64.
// Practical uses: transmitting binary data in HTTP headers/JSON, Basic Auth headers.
func Base64Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// Base64Decode decodes a Base64 string back to plain text.
func Base64Decode(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid Base64 string: %w", err)
	}
	return string(data), nil
}

// FormatJSON pretty-prints and validates a JSON string.
// Practical uses: debugging API responses, making config files more readable.
func FormatJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return string(formatted), nil
}

// RegexMatch finds all strings matching the given regex pattern in the text.
// Practical uses: extracting emails/phone numbers/IPs from logs, validating data formats.
func RegexMatch(pattern, text string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for pattern: %q", pattern), nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pattern : %s\nFound %d match(es):\n%s\n", pattern, len(matches), strings.Repeat("─", 40)))
	for i, m := range matches {
		sb.WriteString(fmt.Sprintf("  [%d] %q\n", i+1, m))
	}
	return sb.String(), nil
}

// WordCount counts characters, words, and lines in the given text.
// Practical uses: checking content length, document statistics.
func WordCount(text string) string {
	lines := strings.Count(text, "\n") + 1
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	charsNoSpace := len([]rune(strings.ReplaceAll(text, " ", "")))
	return fmt.Sprintf(
		"Text statistics:\n%s\n  Characters (with spaces)    : %d\n  Characters (without spaces) : %d\n  Words                       : %d\n  Lines                       : %d",
		strings.Repeat("─", 40), chars, charsNoSpace, words, lines,
	)
}

// TextTransform transforms the text format.
// Practical uses: normalizing variable names, preparing data before storing to DB.
func TextTransform(text, operation string) (string, error) {
	switch strings.ToLower(operation) {
	case "upper":
		return strings.ToUpper(text), nil
	case "lower":
		return strings.ToLower(text), nil
	case "title":
		return toTitleCase(text), nil
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case "trim":
		return strings.TrimSpace(text), nil
	case "snake_case":
		return toSnakeCase(text), nil
	case "camel_case":
		return toCamelCase(text), nil
	default:
		return "", fmt.Errorf(
			"invalid operation: '%s'\nSupported operations: upper, lower, title, reverse, trim, snake_case, camel_case",
			operation,
		)
	}
}

// GetCurrentTime returns the current time in the specified timezone.
// Practical uses: multi-region applications, standardized logging timestamps, regional reports.
func GetCurrentTime(timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf(
			"invalid timezone: '%s'\nValid examples: UTC, America/New_York, Europe/London, Asia/Tokyo",
			timezone,
		)
	}
	now := time.Now().In(loc)
	year, week := now.ISOWeek()
	_, month, day := now.Date()

	return fmt.Sprintf(
		"Current time (%s)\n%s\n  RFC3339        : %s\n  Date           : %02d/%02d/%d\n  Time           : %02d:%02d:%02d\n  Unix Timestamp : %d\n  ISO Week       : Week %d of %d\n  Day of year    : %d/%d",
		timezone,
		strings.Repeat("─", 50),
		now.Format(time.RFC3339),
		day, int(month), now.Year(),
		now.Hour(), now.Minute(), now.Second(),
		now.Unix(),
		week, year,
		now.YearDay(), daysInYear(now.Year()),
	), nil
}

// --- Internal helpers ---

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(w)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			sb.WriteRune('_')
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return strings.ReplaceAll(sb.String(), " ", "_")
}

func toCamelCase(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		result += toTitleCase(w)
	}
	return result
}

func daysInYear(year int) int {
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		return 366
	}
	return 365
}
