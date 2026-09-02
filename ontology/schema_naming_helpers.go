package ontology

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var nonIdentifier = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func CapitalizeWordsManual(input string) string {
	words := strings.Fields(input) // @INFO: Splits by whitespace
	for i, word := range words {
		if word == "" {
			continue
		}
		// @HINT: Decode the first rune (safely handle Unicode characters)
		r, size := utf8.DecodeRuneInString(word)
		if r == utf8.RuneError {
			continue
		}
		// @HINT: Capitalize the first rune and append the rest of the word
		words[i] = string(unicode.ToUpper(r)) + word[size:]
	}
	return strings.Join(words, " ")
}

func NormalizeIdentifier(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	value = nonIdentifier.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")

	if value == "" {
		return "unnamed"
	}

	return value
}

func JunctionTableName(
	source string,
	relationship Relationship,
	index int,
) string {
	return fmt.Sprintf(
		"%s_%s_%s_%d",
		NormalizeIdentifier(source),
		NormalizeIdentifier(relationship.Name),
		NormalizeIdentifier(relationship.Target),
		index,
	)
}
