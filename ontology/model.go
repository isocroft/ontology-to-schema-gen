package ontology

import (
	"fmt"
	"strconv"
	"strings"
)

type Ontology struct {
	Name        string
	Description string

	Entities map[string]*Entity
	Enums    map[string]*Enum

	Warnings []string
}

type Entity struct {
	Name         string
	Description  string
	Properties   []Property
	Relationships []Relationship
}

type Property struct {
	Name        string
	Type        string
	Unique      bool
	Constraints []Constraint
}

type Constraint struct {
	Kind  string
	Value string
}

type Relationship struct {
	Name        string
	Target      string
	Cardinality Cardinality
}

// @HINT: Cardinality represents relationship multiplicity constraints.
type Cardinality uint8

const (
	CardinalityUnknown Cardinality = iota
	CardinalityOne                  // "one" (1)
	CardinalityZeroOrOne            // "zero..one" (0..1)
	CardinalityZeroOrMany           // "zero..many" (0..*)
	CardinalityMany                 // "one..many" (1..*)
)

// @HINT: Checks whether the `Cardinality` value is a recognized/non-unknown state.
func (c Cardinality) IsValid() bool {
	return c > CardinalityUnknown && c <= CardinalityMany
}

// @HINT: Returns the canonical string representation of a `Cardinality`.
func (c Cardinality) String() (string, error) {
	switch c {
	  	case CardinalityOne:
	  		return "one", nil
	  	case CardinalityZeroOrOne:
	  		return "zero..one", nil
	  	case CardinalityZeroOrMany:
	  		return "zero..many", nil
	  	case CardinalityMany:
	  		return "many", nil
	  	default:
	    	if !c.IsValid() {
	        	s := strconv.FormatUint(uint64(c), 10)
	        	return "", fmt.Errorf("invalid cardinality type: %q", s)
	      	}
		  	return "unknown", nil
	}
}

// @HINT: Converts a string into a strongly-typed Cardinality.
func ParseCardinality(s string) (Cardinality, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	  	case "one", "1":
	  		return CardinalityOne, nil
	  	case "zero..one", "0..1", "optional":
	  		return CardinalityZeroOrOne, nil
	  	case "zero..many", "0..*":
	  		return CardinalityZeroOrMany, nil
	  	case "many", "one..many", "1..*":
	  		return CardinalityMany, nil
	  	default:
	  		return CardinalityUnknown, fmt.Errorf("invalid cardinality representation: %q", s)
	}
}

type Enum struct {
	Name   string
	Values []string
}
