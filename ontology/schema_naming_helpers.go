package ontology

import (
	"fmt"
	"regexp"
	"strings"
)

var nonIdentifier = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

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
