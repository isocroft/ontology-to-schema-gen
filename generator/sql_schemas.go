package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/isocroft/ontology-to-schema-gen/ontology"
)

type SQLDialect string

const (
	PostgreSQL SQLDialect = "postgres"
	SQLite     SQLDialect = "sqlite"
	MySQL      SQLDialect = "mysql"
)

func GenerateSQL(
  ontologyName string,
	o *ontology.Ontology,
	dialect SQLDialect,
) string {
	var b strings.Builder

	fmt.Fprintf(
		&b,
		fmt.Sprintf("-- %s Ontology SQL\n", ontology.CapitalizeWordsManual(strings.ToLower(ontologyName))) + "-- Dialect: %s\n\n",
		dialect,
	)

	entityNames := make([]string, 0, len(o.Entities))

	for name := range o.Entities {
		entityNames = append(entityNames, name)
	}

	sort.Strings(entityNames)

	// @HINT: Base tables.
	for _, name := range entityNames {
		writeEntityTable(
			&b,
			o,
			o.Entities[name],
			dialect,
		)

		b.WriteString("\n")
	}

	// @HINT: Junction tables.
	for _, name := range entityNames {
		entity := o.Entities[name]

		manyIndex := 0

		for _, relationship := range entity.Relationships {
			if relationship.Cardinality != ontology.CardinalityMany {
				continue
			}

			manyIndex++

			writeJunctionTable(
				&b,
				entity,
				relationship,
				manyIndex,
				dialect,
			)

			b.WriteString("\n")
		}
	}

	return b.String()
}

func writeEntityTable(
	b *strings.Builder,
	o *ontology.Ontology,
	entity *ontology.Entity,
	dialect SQLDialect,
) {
	table := ontology.NormalizeIdentifier(entity.Name)

	fmt.Fprintf(
		b,
		"CREATE TABLE %s (\n",
		quoteIdentifier(table, dialect),
	)

	columns := []string{}

	for _, property := range entity.Properties {
		column := propertyColumn(
			property,
			o,
			dialect,
		)

		columns = append(columns, column)
	}

	// Singular relationships become FK columns.
	for _, relationship := range entity.Relationships {
		if relationship.Cardinality == ontology.CardinalityMany {
			continue
		}

		target := ontology.NormalizeIdentifier(
			relationship.Target,
		)

		columnName := ontology.NormalizeIdentifier(
			relationship.Name,
		) + "_" + target + "_id"

		sqlType := primaryKeyType(dialect)

		nullable := relationship.Cardinality == ontology.CardinalityZeroOrOne

		column := fmt.Sprintf(
			"    %s %s",
			quoteIdentifier(columnName, dialect),
			sqlType,
		)

		if nullable {
			column += " NULL"
		} else {
			column += " NOT NULL"
		}

		columns = append(columns, column)
	}

	// Ensure a comma between column definitions and constraints.
	for i, column := range columns {
		if i > 0 {
			b.WriteString(",\n")
		}

		b.WriteString(column)
	}

	// Primary key.
	if len(columns) > 0 {
		b.WriteString(",\n")
	}

	b.WriteString(
		"    PRIMARY KEY (" +
			quoteIdentifier("id", dialect) +
			")",
	)

	// Foreign keys.
	for _, relationship := range entity.Relationships {
		if relationship.Cardinality == ontology.CardinalityMany {
			continue
		}

		targetTable := ontology.NormalizeIdentifier(
			relationship.Target,
		)

		columnName :=
			ontology.NormalizeIdentifier(
				relationship.Name,
			) +
				"_" +
				targetTable +
				"_id"

		b.WriteString(",\n")

		fmt.Fprintf(
			b,
			"    FOREIGN KEY (%s) REFERENCES %s (%s)",
			quoteIdentifier(columnName, dialect),
			quoteIdentifier(targetTable, dialect),
			quoteIdentifier("id", dialect),
		)
	}

	b.WriteString("\n);\n")

	// @HINT: Unique constraints.
	for _, property := range entity.Properties {
		if !property.Unique {
			continue
		}

		fmt.Fprintf(
			b,
			"CREATE UNIQUE INDEX %s ON %s (%s);\n",
			quoteIdentifier(
				table+"_"+ontology.NormalizeIdentifier(property.Name)+"_uk",
				dialect,
			),
			quoteIdentifier(table, dialect),
			quoteIdentifier(
				ontology.NormalizeIdentifier(property.Name),
				dialect,
			),
		)
	}
}

func writeJunctionTable(
	b *strings.Builder,
	source *ontology.Entity,
	relationship ontology.Relationship,
	index int,
	dialect SQLDialect,
) {
	table := ontology.JunctionTableName(
		source.Name,
		relationship,
		index,
	)

	sourceTable := ontology.NormalizeIdentifier(
		source.Name,
	)

	targetTable := ontology.NormalizeIdentifier(
		relationship.Target,
	)

	sourceColumn := sourceTable + "_id"
	targetColumn := targetTable + "_id"

	fmt.Fprintf(
		b,
		"CREATE TABLE %s (\n",
		quoteIdentifier(table, dialect),
	)

	fmt.Fprintf(
		b,
		"    %s %s NOT NULL,\n",
		quoteIdentifier(sourceColumn, dialect),
		primaryKeyType(dialect),
	)

	fmt.Fprintf(
		b,
		"    %s %s NOT NULL,\n",
		quoteIdentifier(targetColumn, dialect),
		primaryKeyType(dialect),
	)

	fmt.Fprintf(
		b,
		"    PRIMARY KEY (%s, %s),\n",
		quoteIdentifier(sourceColumn, dialect),
		quoteIdentifier(targetColumn, dialect),
	)

	fmt.Fprintf(
		b,
		"    FOREIGN KEY (%s) REFERENCES %s (%s),\n",
		quoteIdentifier(sourceColumn, dialect),
		quoteIdentifier(sourceTable, dialect),
		quoteIdentifier("id", dialect),
	)

	fmt.Fprintf(
		b,
		"    FOREIGN KEY (%s) REFERENCES %s (%s)\n",
		quoteIdentifier(targetColumn, dialect),
		quoteIdentifier(targetTable, dialect),
		quoteIdentifier("id", dialect),
	)

	b.WriteString(");\n")
}

func propertyColumn(
	property ontology.Property,
	o *ontology.Ontology,
	dialect SQLDialect,
) string {
	columnName := ontology.NormalizeIdentifier(
		property.Name,
	)

	sqlType := sqlType(
		property.Type,
		o,
		dialect,
	)

	column := fmt.Sprintf(
		"    %s %s",
		quoteIdentifier(columnName, dialect),
		sqlType,
	)

	if property.Name == "id" {
		column += " NOT NULL"
	}

	return column
}

func sqlType(
	typeName string,
	o *ontology.Ontology,
	dialect SQLDialect,
) string {
	if _, isEnum := o.Enums[typeName]; isEnum {
		switch dialect {
		case PostgreSQL:
			return "TEXT"

		case MySQL:
			return "VARCHAR(64)"

		case SQLite:
			return "TEXT"
		}
	}

	switch strings.ToLower(typeName) {
	case "uuid":
		switch dialect {
		case PostgreSQL:
			return "UUID"

		case MySQL:
			return "BINARY(16)"

		case SQLite:
			return "TEXT"
		}

	case "datetime":
		switch dialect {
		case PostgreSQL:
			return "TIMESTAMPTZ"

		case MySQL:
			return "TIMESTAMP"

		case SQLite:
			return "TEXT"
		}

	case "string":
		switch dialect {
		case PostgreSQL:
			return "TEXT"

		case MySQL:
			return "VARCHAR(255)"

		case SQLite:
			return "TEXT"
		}

	case "integer", "int":
		return "INTEGER"

	case "int32":
		return "INTEGER"

	case "int64":
		return "BIGINT"

	case "float":
		return "REAL"

	case "double":
		return "DOUBLE PRECISION"

	case "boolean", "bool":
		switch dialect {
		case PostgreSQL:
			return "BOOLEAN"

		case MySQL:
			return "TINYINT(1)"

		case SQLite:
			return "INTEGER"
		}
	}

	return "TEXT"
}

func primaryKeyType(dialect SQLDialect) string {
	switch dialect {
	case PostgreSQL:
		return "UUID"

	case MySQL:
		return "BINARY(16)" // @NOTE: For inserts use `UUID_TO_BIN(...)`

	case SQLite:
		return "TEXT"

	default:
		return "BIGINT"
	}
}

func quoteIdentifier(
	value string,
	dialect SQLDialect,
) string {
	switch dialect {
	case MySQL:
		return "`" + value + "`"

	default:
		return `"` + value + `"`
	}
}
