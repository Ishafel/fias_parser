package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fias_parser/pkg/schema"
	"fias_parser/pkg/xmlstream"
)

func main() {
	schemaDir := flag.String("schema-dir", "gar_schemas", "Directory with GAR XSD schemas")
	xmlPath := flag.String("xml", "", "Path to XML file or directory to parse")
	elementName := flag.String("element", "", "Name of the element to stream (defaults to first child of root)")
	expectedCount := flag.Int("expected-count", -1, "Expected number of records per XML file; if provided, mismatches are logged")
	warnLog := flag.String("warn-log", "validation.log", "Path to append validation warnings")
	flag.Parse()

	if *xmlPath == "" {
		fmt.Fprintln(os.Stderr, "--xml is required")
		os.Exit(1)
	}

	schemas, err := schema.LoadSchemas(*schemaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load schemas: %v\n", err)
		os.Exit(1)
	}

	files, err := collectXMLFiles(*xmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect xml files: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		xmlRoot, err := xmlstream.DetectXMLRoot(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "detect xml root in %s: %v\n", file, err)
			os.Exit(1)
		}

		currentSchema, ok := schemas[xmlRoot]
		if !ok {
			fmt.Fprintf(os.Stderr, "no schema found for root element %q in %s\n", xmlRoot, file)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Using schema %s for root element %s in %s\n", currentSchema.Path, currentSchema.RootElement, file)

		count, err := xmlstream.StreamElements(file, *elementName, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream xml from %s: %v\n", file, err)
			os.Exit(1)
		}

		if *expectedCount >= 0 && count != *expectedCount {
			msg := fmt.Sprintf("expected %d records but found %d in %s", *expectedCount, count, file)
			if err := appendWarning(*warnLog, msg); err != nil {
				fmt.Fprintf(os.Stderr, "log warning: %v\n", err)
			}
			fmt.Fprintln(os.Stderr, "warning:", msg)
		}
	}
}

func appendWarning(path string, msg string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, msg); err != nil {
		return err
	}

	return nil
}

func collectXMLFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), ".xml") {
			files = append(files, filepath.Join(path, name))
		}
	}

	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no xml files found in %s", path)
	}

	return files, nil
}
