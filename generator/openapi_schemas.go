package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/isocroft/ontology-to-schema-gen/ontology"
)

func GenerateOpenAPI(ontologyName string, o *ontology.Ontology) string {
	var b strings.Builder

	b.WriteString("openapi: 3.0.3\n")
	b.WriteString("info:\n")
	b.WriteString(fmt.Sprintf("  title: %s API\n", ontology.CapitalizeWordsManual(strings.ToLower(ontologyName))))
	b.WriteString("  version: 1.0.0\n")
	b.WriteString(fmt.Sprintf("  description: Generated from %s ontology.\n", strings.ToLower(ontologyName)))

	b.WriteString("components:\n")
	b.WriteString("  schemas:\n")

	entityNames := make([]string, 0, len(o.Entities))

	for name := range o.Entities {
		entityNames = append(entityNames, name)
	}

	sort.Strings(entityNames)

	for _, name := range entityNames {
		entity := o.Entities[name]

		fmt.Fprintf(
			&b,
			"    %s:\n",
			entity.Name,
		)

		b.WriteString("      type: object\n")

		if entity.Description != "" {
			fmt.Fprintf(
				&b,
				"      description: %q\n",
				entity.Description,
			)
		}

		b.WriteString("      properties:\n")

		for _, property := range entity.Properties {
			fmt.Fprintf(
				&b,
				"        %s:\n",
				property.Name,
			)

			writeOpenAPIType(
				&b,
				property.Type,
				o,
				property.Constraints,
			)

			if property.Unique {
				b.WriteString("          x-unique: true\n")
			}
		}

		for _, relationship := range entity.Relationships {
			writeRelationshipSchema(
				&b,
				relationship,
			)
		}
	}

	enumNames := make([]string, 0, len(o.Enums))

	for name := range o.Enums {
		enumNames = append(enumNames, name)
	}

	sort.Strings(enumNames)

	for _, name := range enumNames {
		enum := o.Enums[name]

		fmt.Fprintf(
			&b,
			"    %s:\n",
			enum.Name,
		)

		b.WriteString("      type: string\n")
		b.WriteString("      enum:\n")

		for _, value := range enum.Values {
			fmt.Fprintf(
				&b,
				"        - %s\n",
				value,
			)
		}
	}

	return b.String()
}

func writeOpenAPIType(
	b *strings.Builder,
	typeName string,
	o *ontology.Ontology,
	constraints []ontology.Constraint,
) {
	if enum, exists := o.Enums[typeName]; exists {
		b.WriteString("          type: string\n")
		fmt.Fprintf(
			b,
			"          $ref: '#/components/schemas/%s'\n",
			enum.Name,
		)
		return
	}

	switch strings.ToLower(typeName) {
	case "uuid":
		b.WriteString("          type: string\n")
		b.WriteString("          format: uuid\n")

	case "datetime":
		b.WriteString("          type: string\n")
		b.WriteString("          format: date-time\n")

	case "string":
		b.WriteString("          type: string\n")

	case "integer", "int":
		b.WriteString("          type: integer\n")

	case "int32":
		b.WriteString("          type: integer\n")
		b.WriteString("          format: int32\n")

	case "int64":
		b.WriteString("          type: integer\n")
		b.WriteString("          format: int64\n")

	case "float":
		b.WriteString("          type: number\n")
		b.WriteString("          format: float\n")

	case "double":
		b.WriteString("          type: number\n")
		b.WriteString("          format: double\n")

	case "boolean", "bool":
		b.WriteString("          type: boolean\n")

	default:
		b.WriteString("          type: string\n")
		b.WriteString(
			"          x-ontology-type: " + typeName + "\n",
		)
	}

	for _, constraint := range constraints {
		switch constraint.Kind {
		case "REGEX":
			if constraint.Value != "" {
				fmt.Fprintf(
					b,
					"          pattern: %q\n",
					constraint.Value,
				)
			}

		case "URI":
			if strings.EqualFold(
				constraint.Value,
				"email",
			) {
				b.WriteString(
					"          format: email\n",
				)
			}

		default:
			if constraint.Kind != "" {
				fmt.Fprintf(
					b,
					"          x-constraint-%s: %q\n",
					ontology.NormalizeIdentifier(
						constraint.Kind,
					),
					constraint.Value,
				)
			}
		}
	}
}

func writeRelationshipSchema(
	b *strings.Builder,
	r ontology.Relationship,
) {
	propertyName := ontology.NormalizeIdentifier(r.Name)

	// Duplicate relationship names are possible in the ontology.
	// The parser preserves them, so the OpenAPI property name must
	// also be unique.
	if r.Target != "" {
		propertyName += "_" + ontology.NormalizeIdentifier(r.Target)
	}

	fmt.Fprintf(
		b,
		"        %s:\n",
		propertyName,
	)

	switch r.Cardinality {
	case ontology.CardinalityMany:
		b.WriteString("          type: array\n")
		b.WriteString("          items:\n")
		fmt.Fprintf(
			b,
			"            $ref: '#/components/schemas/%s'\n",
			r.Target,
		)

	case ontology.CardinalityOne:
		fmt.Fprintf(
			b,
			"          $ref: '#/components/schemas/%s'\n",
			r.Target,
		)

	case ontology.CardinalityZeroOrOne:
		b.WriteString("          oneOf:\n")
		fmt.Fprintf(
			b,
			"            - $ref: '#/components/schemas/%s'\n",
			r.Target,
		)
		b.WriteString("          nullable: true\n")

	default:
		fmt.Fprintf(
			b,
			"          $ref: '#/components/schemas/%s'\n",
			r.Target,
		)
	}
}
