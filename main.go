package main

import (
	"fmt"
	"os"

	"github.com/isocroft/ontology-to-schema-gen/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error |> \r\n", err)
		os.Exit(1)
	}
}
