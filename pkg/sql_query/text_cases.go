package sql_query

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}

	builder := strings.Builder{}
	builder.WriteString(strings.ToLower(parts[0])) // bagian pertama lowercase

	titleCaser := cases.Title(language.English)
	for _, part := range parts[1:] {
		builder.WriteString(titleCaser.String(part)) // kapitalisasi dengan Unicode-aware
	}

	return builder.String()
}

func CamelToSnake(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func PascalToCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
