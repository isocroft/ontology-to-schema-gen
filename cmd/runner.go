package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isocroft/ontology-to-schema-gen/generator"
	"github.com/isocroft/ontology-to-schema-gen/ontology"
)

func Run() error {
	input := flag.String(
		"input",
		".",
		"Directory containing the ontology YAML files",
	)

	output := flag.String(
		"output",
		"./generated",
		"Output directory",
	)

	ontologyName := flag.String(
		"name",
		"MY",
		"Schema name",
	)

	openapi := flag.Bool(
		"openapi",
		true,
		"Generate OpenAPI 3 schema",
	)

	sql := flag.Bool(
		"sql",
		false,
		"Generate SQL schemas",
	)

	postgres := flag.Bool(
		"postgres",
		false,
		"Generate PostgreSQL schema",
	)

	sqlite := flag.Bool(
		"sqlite",
		false,
		"Generate SQLite schema",
	)

	mysql := flag.Bool(
		"mysql",
		false,
		"Generate MySQL schema",
	)

	showWarnings := flag.Bool(
		"warnings",
		true,
		"Display YAML parser warnings",
	)

	dumpModel := flag.Bool(
		"dump-model",
		false,
		"Dump the normalized ontology model as JSON",
	)

	flag.Parse()

	if err := os.MkdirAll(*output, 0o755); err != nil {
		return fmt.Errorf(
			"creating output directory failed; reason: %w",
			err,
		)
	}

	parser := ontology.NewYAMLParser()

	model, err := parser.ParseDirectory(*input)
	if err != nil {
		return err
	}

	if *showWarnings {
		for _, warning := range model.Warnings {
			fmt.Printf("WARNING: %s\n", warning)
		}
	}

	if *dumpModel {
		if err := dumpNormalizedModel(
			model,
			filepath.Join(*output, "ontology-model.json"),
		); err != nil {
			return err
		}
	}

	if *openapi {
		path := filepath.Join(
			*output,
			"openapi.yaml",
		)

		if err := os.WriteFile(
			path,
			[]byte(generator.GenerateOpenAPI(*ontologyName, model)),
			0o644,
		); err != nil {
			return fmt.Errorf(
				"making OpenAPI schema failed; %w",
				err,
			)
		}

		fmt.Println("generated:", path)
	}

	if *sql {
		if *postgres {
			if err := writeSQL(
				model,
				generator.PostgreSQL,
				*output,
				"postgres.sql",
			); err != nil {
				return err
			}
		}

		if *sqlite {
			if err := writeSQL(
				model,
				generator.SQLite,
				*output,
				"sqlite.sql",
			); err != nil {
				return err
			}
		}

		if *mysql {
			if err := writeSQL(
				model,
				generator.MySQL,
				*output,
				"mysql.sql",
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeSQL(
	model *ontology.Ontology,
	dialect generator.SQLDialect,
	output string,
	filename string,
) error {
	path := filepath.Join(output, filename)

	if err := os.WriteFile(
		path,
		[]byte(generator.GenerateSQL(model.Name, model, dialect)),
		0o644,
	); err != nil {
		return fmt.Errorf(
			"making %s schema failed; %w",
			dialect,
			err,
		)
	}

	fmt.Println("generated:", path)

	return nil
}

func dumpNormalizedModel(
	model *ontology.Ontology,
	path string,
) error {
	data, err := json.MarshalIndent(
		model,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"making normalized model failed; %w",
			err,
		)
	}

	if err := os.WriteFile(
		path,
		data,
		0o644,
	); err != nil {
		return fmt.Errorf(
			"making normalized model failed; %w",
			err,
		)
	}

	fmt.Println("generated:", path)

	return nil
}
