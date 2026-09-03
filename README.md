A simple yet robust Golang program that reads a set of ontology references for any domain defined in YAML and converts them to schema definitions.

## Getting Started

You can build the project like so:

```bash
$ go mod tidy
$ go build -o ontology-to-schema-gen .
```

>Then, move to the directory just above the directory where the ontolgy YAML files are located and run any of the commands below in a terminal.

```bash
$ ./ontology-to-schema-gen \
    -input ./fleet-management-ontology \
    -output ./files
```

```bash
$ ./ontology-to-schema-gen \
    -input ./fleet-management-ontology \
    -output ./files \
    -sql=true \
    -postgres=true \
    -sqlite=true \
    -mysql=true
```

```bash
$ ./ontology-to-schema-gen \
    -input ./fleet-management-ontology \
    -output ./files \
    -dump-model=true
```

## Generated Output

The command above releases any or all of the files listed below.

>Default folder name from optional command-line arguments (**generated**):
```
generated/
├── openapi.yaml
├── postgres.sql
├── sqlite.sql
└── mysql.sql
```

>Custom folder name set on command-line (**files**):
```
files/
├── ontology-model.json
├── openapi.yaml
├── postgres.sql
├── sqlite.sql
└── mysql.sql
```

## Step-by-Step Automation Pipeline for Development Workflow.

1. **Map Ontology to OpenAPI YAML**

**ontology-to-schema-gen** makes use of the ontology provided on the command-line (e.g., entities, properties, enums, relationships) to programmatically convert these ontology structures into valid OpenAPI v3.0.3 component schemas (components/schemas) saved as an `openapi.yaml` file.

2. **Generate TypeScript Types**

Subsequently, the CLI app from [**OpenAPI TypeScript**](https://www.npmjs.com/package/openapi-typescript) can be employed to read your generated OpenAPI YAML file and output typed interfaces. Run the transformation command in your terminal or build script as follows:

```bash
$ npx openapi-typescript ./files/openapi.yaml -o generated-models.ts
```

Use the code above with some caution.

3. **Automate with File Watchers or CI/CD**

   - **Local Development**: Use a file watcher like `chokidar-cli` or `nodemon` to watch your ontology YAML folder. Trigger your custom script and the _openapi-typescript_ compiler automatically whenever a file changes.
   
   - **CI/CD Pipeline**: Add a build step in GitHub Actions or your local `package.json` scripts to run the conversion before compilation or testing, ensuring your TypeScript definitions never drift from your source ontology.

### TODOS

1. The generator does not currently try to infer inverse relationships. It treats each YAML relationship as an independently declared relationship. A future version should therefore add an explicit relationship reconciliation phase. That would allow the SQL generator to produce a much more semantically accurate relational model rather than mechanically translating each relationship.
2. The generator does not currently convert the `JSON` constraint into a structured OpenAPI object or JSON/JSONB SQL column. It retains the constraint as ontology metadata instead. There would be a need to implement a conversion step for this constraint type. This also applies to the `E.164:phone` constraint type as well.
