package main

import (
	"flag"
	"fmt"
	"os"

	"fias_parser/pkg/schema"
	"fias_parser/pkg/xmlstream"
)

func main() {
	schemaDir := flag.String("schema-dir", "gar_schemas", "Directory with GAR XSD schemas")
	xmlFile := flag.String("xml", "", "Path to XML file to parse")
	elementName := flag.String("element", "", "Name of the element to stream (defaults to first child of root)")
	flag.Parse()

	if *xmlFile == "" {
		fmt.Fprintln(os.Stderr, "--xml is required")
		os.Exit(1)
	}

	schemas, err := schema.LoadSchemas(*schemaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load schemas: %v\n", err)
		os.Exit(1)
	}

	xmlRoot, err := xmlstream.DetectXMLRoot(*xmlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect xml root: %v\n", err)
		os.Exit(1)
	}

	currentSchema, ok := schemas[xmlRoot]
	if !ok {
		fmt.Fprintf(os.Stderr, "no schema found for root element %q\n", xmlRoot)
		os.Exit(1)
	}

	fmt.Printf("Using schema %s for root element %s\n", currentSchema.Path, currentSchema.RootElement)

	if err := xmlstream.StreamElements(*xmlFile, *elementName, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "stream xml: %v\n", err)
		os.Exit(1)
	}
}
