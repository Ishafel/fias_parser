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

		targetElement, expected, err := xmlstream.CountElements(file, *elementName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "count elements in %s: %v\n", file, err)
			os.Exit(1)
		}

		requiredAttrs := currentSchema.RequiredAttributes[targetElement]

		result, err := xmlstream.StreamElements(file, targetElement, expected, requiredAttrs, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream xml from %s: %v\n", file, err)
			os.Exit(1)
		}

		if result.Processed != result.Expected || len(result.Skipped) > 0 {
			if err := logValidation(*warnLog, file, result); err != nil {
				fmt.Fprintf(os.Stderr, "log warning: %v\n", err)
			}
			fmt.Fprintf(os.Stderr, "validation warning: expected %d records but processed %d in %s\n", result.Expected, result.Processed, file)
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

func logValidation(path string, file string, result xmlstream.StreamResult) error {
	summary := fmt.Sprintf("expected %d records but processed %d in %s", result.Expected, result.Processed, file)
	if err := appendWarning(path, summary); err != nil {
		return err
	}

	for _, skipped := range result.Skipped {
		msg := fmt.Sprintf("%s: skipped record #%d at byte %d for element %s: %s", file, skipped.Index, skipped.ByteOffset, skipped.Element, skipped.Error)
		if err := appendWarning(path, msg); err != nil {
			return err
		}
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
