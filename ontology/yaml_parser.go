package ontology

import (
	"fmt"
	"os"
  "path/filepath"
  "strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type YAMLParser struct {
	Warnings []string
}

type MapEntry struct {
	Key   *yaml.Node
	Value *yaml.Node
}

func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

func (p *YAMLParser) parseFile(path string, dir string) (*yaml.Node, error) {
  errMsg_Prefix := fmt.Sprintf("unable to parse file (%s) in directory (%q); ", path, dir)
  /*!
    @TODO:
  
    Using `os.ReadFile(...)` for now. Will need to 
    change to something else for larger YAML files.
  */
	data, err := os.ReadFile(path)
  
	if err != nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
	}

	var root yaml.Node

	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
	}

	if len(root.Content) == 0 {
    err := fmt.Errorf("empty YAML document found")
		return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
	}

	return root.Content[0], nil
}

func (p *YAMLParser) ParseDirectory(dir string) (*Ontology, error) {
  exePath, err := os.Executable()
  
	if err != nil {
		panic(err)
	}

  rootdir := filepath.Dir(exePath)
  
  errMsg_Prefix := fmt.Sprintf("failed to complete parsing for ontology definitions generated from (%s); ", rootdir)
	ontology := &Ontology{
		Entities: make(map[string]*Entity),
		Enums:    make(map[string]*Enum),
	}

	files, err := filepath.Glob(filepath.Join(dir, "**", "*.yml"))
  
	if err != nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
	}

	/*!
    @INFO:
    
    `filepath.Glob` doesn't recursively expand `**`.
    Therefore, use the filesystem walker instead.
  */
	files, err = yamlFiles(dir)
  
	if err != nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
	}

	for _, path := range files {
		root, err := p.parseFile(path, dir)
    
		if err != nil {
			return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
		}

		switch {
      case MappingValue(root, "entity") != nil:
        entity, err := p.parseEntity(root, path)
        if err != nil {
          return nil, err
        }
  
        if _, exists := ontology.Entities[entity.Name]; exists {
          err := fmt.Errorf(
            "duplicate entity definition '%s'",
            entity.Name,
          )
          return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
        }
  
        ontology.Entities[entity.Name] = entity
  
      case MappingValue(root, "enum") != nil:
        enum, err := p.parseEnum(root)
        if err != nil {
          return nil, err
        }
  
        if _, exists := ontology.Enums[enum.Name]; exists {
          err := fmt.Errorf(
            "duplicate enum definition '%s'",
            enum.Name,
          )
          return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
        }
  
        ontology.Enums[enum.Name] = enum
		}
	}

	ontology.Warnings = append(ontology.Warnings, p.Warnings...)

	return ontology, nil
}

func (p *YAMLParser) parseEntity(root *yaml.Node, path string) (*Entity, error) {
  errMsg_Prefix := "issue while attempting to parse entity; "
	node := MappingValue(root, "entity")

	if node == nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: missing entity")
	}

	entity := &Entity{
		Name:         StringValue(MappingValue(node, "name")),
		Description:  StringValue(MappingValue(node, "description")),
		Properties:   []Property{},
		Relationships: []Relationship{},
	}

	if entity.Name == "" {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: entity name is required")
	}

	properties := MappingValue(node, "properties")
  relationships := MappingValue(node, "relationships")

  if properties == nil || relationships == nil {
    return entity, nil
  }


  for i := 0; i < len(properties.Content); i += 2 {
    nameNode := properties.Content[i]
    valueNode := properties.Content[i+1]

    prop, err := p.parseEntityProperty(
      nameNode.Value,
      valueNode
    )

    if err != nil {
      return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
    }

    entity.Properties = append(entity.Properties, prop)
  }
	

	if relationships != nil {
		seen := make(map[string]int)

		/*!
      @NOTE:
      
      DO NOT convert this mapping into `map[string]yaml.Node`.
		  This is so that duplicate relationship keys always survive.
    */
		for i := 0; i < len(relationships.Content); i += 2 {
			nameNode := relationships.Content[i]
			valueNode := relationships.Content[i+1]

			name := nameNode.Value

			seen[name]++

			if seen[name] > 1 {
				p.Warnings = append(
					p.Warnings,
					fmt.Sprintf(
						"%s: duplicate relationship key '%q' preserved as occurrence #%d",
						path,
						name,
						seen[name],
					),
				)
			}

			relationship, err := p.parseEntityRelationship(
				name,
				valueNode
			)

			if err != nil {
				return nil, fmt.Errorf(errMsg_Prefix + "reason: %w", err)
			}

			entity.Relationships = append(
				entity.Relationships,
				relationship,
			)
		}
	}

	return entity, nil
}

func (p *YAMLParser) parseEntityProperty(
	name string,
	node *yaml.Node
) (Property, error) {
  errMsg_Prefix := "cannot process entity property; "
	prop := Property{
		Name:        name,
		Type:        StringValue(MappingValue(node, "type")),
		Unique:      BoolValue(MappingValue(node, "unique")),
		Constraints: []Constraint{},
	}

	if prop.Type == "" {
		return prop, fmt.Errorf(
			errMsg_Prefix + fmt.Sprintf("reason: property '%s' has no type",
			name)
		)
	}

	constraints := MappingValue(node, "constraint")

	if constraints != nil {
    if constraints.Kind != yaml.SequenceNode {
      return prop, fmt.Errorf(errMsg_Prefix + "reason: invalid value kind for constraint")
    }
    
    if constraints.Kind == yaml.SequenceNode {
  		for _, item := range constraints.Content {
  			value := item.Value
  
  			kind := value
  			constraintValue := ""
  
  			if idx := strings.IndexByte(value, ':'); idx >= 0 {
  				kind = value[:idx]
  				constraintValue = value[idx+1:]
  			}
  
  			prop.Constraints = append(
  				prop.Constraints,
  				Constraint{
  					Kind:  strings.TrimSpace(kind),
  					Value: strings.TrimSpace(constraintValue),
  				},
  			)
  		}
    }
	}

	return prop, nil
}

func (p *YAMLParser) parseEntityRelationship(
	name string,
	node *yaml.Node
) (Relationship, error) {
  errMsg_Prefix := "cannot process entity relationship; "
	target := StringValue(MappingValue(node, "target"))
	cardinality := StringValue(MappingValue(node, "cardinality"))

	if target == "" {
		return Relationship{}, fmt.Errorf(
			errMsg_Prefix + fmt.Sprintf("reason: relationship '%s' has no target",
			name)
		)
	}

  c, err := ParseCardinality(cardinality)

  if err != nil {
    return Relationship{}, fmt.Errorf(
      errMsg_Prefix + "reason: %w for relationship",
      err
    )
  }

	return Relationship{
		Name:        name,
		Target:      target,
		Cardinality: c,
	}, nil
}

func (p *YAMLParser) parseEnum(
	root *yaml.Node,
) (*Enum, error) {
  errMsg_Prefix := "cannot process enum value; "
	node := MappingValue(root, "enum")

	if node == nil {
		return nil, fmt.Errorf(errMsg_Prefix + "reason: missing enum")
	}

	name := StringValue(MappingValue(node, "name"))

	if name == "" {
		return nil, fmt.Errorf(
			errMsg_Prefix + "reason: enum name is required",
		)
	}

	values := SequenceStrings(
		MappingValue(node, "values"),
	)

	if len(values) == 0 {
		return nil, fmt.Errorf(
			errMsg_Prefix + fmt.Sprintf("reason: enum '%s' has no values",
			name)
		)
	}

	return &Enum{
		Name:   name,
		Values: values,
	}, nil
}

func MappingEntries(node *yaml.Node) []MapEntry {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

  /*!
    @HINT:
    
    Single the key-value pairs number half of `node.Content`,
    makes a slice half the size of `node.Content`
  */
	entries := make([]MapEntry, 0, len(node.Content)/2)
  
	for i := 0; i < len(node.Content); i += 2 {
		entries = append(entries, MapEntry{
			Key:   node.Content[i],
			Value: node.Content[i+1],
		})
	}
  
	return entries
}

func MappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]

		if k.Value == key {
			return v
		}
	}

	return nil
}

func MappingValues(node *yaml.Node, key string) []*yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	result := make([]*yaml.Node, 0)

	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]

		if k.Value == key {
			result = append(result, v)
		}
	}

	return result
}

func StringValue(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}

	return strings.TrimSpace(node.Value)
}

func BoolValue(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}

	/*!
    @HINT:
    
    strconv.ParseBool handles "1", "t", "T", "TRUE", 
    "true", "True", "0", "f", "F", "FALSE", "false", 
    "False"
  */
	val, err := strconv.ParseBool(strings.TrimSpace(node.Value))
  
	if err == nil {
		return val
	}

	/*!
    @HINT:
    
    Handle additional standard YAML 1.1 boolean 
    representations ("yes", "no", "on", "off")
  */
	switch strings.ToLower(strings.TrimSpace(node.Value)) {
  	case "yes", "on":
  		return true
  	default:
  		return false
	}
}

func SequenceStrings(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}

	var values := make([]string, 0, len(node.Content))

	for _, item := range node.Content {
		values = append(values, item.Value)
	}

	return values
}
