A simple yet robust Golang program that reads a set of ontology references for any domain defined in YAML and converts them to schema definitions.

## Getting Started

You can build the project like so:

```bash
$ go mod tidy
$ go build -o ontology-to-schema-gen .
```

>Then, move to the directory just above the directory where the ontolgy YAML files are located.

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

### TODOS

1. The generator does not currently try to infer inverse relationships. It treats each YAML relationship as an independently declared relationship. A future version should therefore add an explicit relationship reconciliation phase. That would allow the SQL generator to produce a much more semantically accurate relational model rather than mechanically translating each relationship..
2. The generator does not currently convert the `JSON` constraint into a structured OpenAPI object or JSON/JSONB SQL column. It retains the constraint as ontology metadata instead. There would be a need to implement a conversion step for this constraint type. This also applies to the `E.164:phone` constraint type as well.
