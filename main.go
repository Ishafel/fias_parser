package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type schemaInfo struct {
	Path        string
	RootElement string
}

type record struct {
	Element    string            `json:"element"`
	Attributes map[string]string `json:"attributes"`
	Content    string            `json:"content,omitempty"`
}

func main() {
	schemaDir := flag.String("schema-dir", "gar_schemas", "Directory with GAR XSD schemas")
	xmlFile := flag.String("xml", "", "Path to XML file to parse")
	elementName := flag.String("element", "", "Name of the element to stream (defaults to first child of root)")
	flag.Parse()

	if *xmlFile == "" {
		fmt.Fprintln(os.Stderr, "--xml is required")
		os.Exit(1)
	}

	schemas, err := loadSchemas(*schemaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load schemas: %v\n", err)
		os.Exit(1)
	}

	xmlRoot, err := detectXMLRoot(*xmlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "detect xml root: %v\n", err)
		os.Exit(1)
	}

	schema, ok := schemas[xmlRoot]
	if !ok {
		fmt.Fprintf(os.Stderr, "no schema found for root element %q\n", xmlRoot)
		os.Exit(1)
	}

	fmt.Printf("Using schema %s for root element %s\n", schema.Path, schema.RootElement)

	err = streamElements(*xmlFile, *elementName, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stream xml: %v\n", err)
		os.Exit(1)
	}
}

func loadSchemas(dir string) (map[string]schemaInfo, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.xsd"))
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]schemaInfo)
	for _, path := range matches {
		root, err := detectXSDRoot(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		schemas[root] = schemaInfo{Path: path, RootElement: root}
	}

	if len(schemas) == 0 {
		return nil, errors.New("no schemas found")
	}

	return schemas, nil
}

func detectXSDRoot(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := xml.NewDecoder(bufio.NewReader(f))
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if isElement(start, "element") {
			for _, attr := range start.Attr {
				if attr.Name.Local == "name" {
					return attr.Value, nil
				}
			}
		}
	}

	return "", errors.New("root element not found")
}

func detectXMLRoot(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := xml.NewDecoder(bufio.NewReader(f))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		return start.Name.Local, nil
	}
}

func streamElements(path string, elementName string, out io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(out)
	dec := xml.NewDecoder(bufio.NewReader(f))

	depth := 0
	target := elementName

	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				continue
			}

			if target == "" && depth == 2 {
				target = t.Name.Local
			}

			if depth == 2 && t.Name.Local == target {
				rec, err := buildRecord(dec, t)
				if err != nil {
					return err
				}
				if err := enc.Encode(rec); err != nil {
					return err
				}
				// The matching end element was already consumed inside buildRecord,
				// so compensate for the extra depth increase.
				depth--
				continue
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
			}
		}
	}
}

func buildRecord(dec *xml.Decoder, start xml.StartElement) (record, error) {
	attrs := make(map[string]string, len(start.Attr))
	for _, attr := range start.Attr {
		attrs[attr.Name.Local] = attr.Value
	}

	var content strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return record{}, err
		}

		switch t := tok.(type) {
		case xml.CharData:
			data := strings.TrimSpace(string(t))
			if data != "" {
				if content.Len() > 0 {
					content.WriteRune(' ')
				}
				content.WriteString(data)
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return record{
					Element:    start.Name.Local,
					Attributes: attrs,
					Content:    content.String(),
				}, nil
			}
		case xml.StartElement:
			if err := dec.Skip(); err != nil {
				return record{}, err
			}
		}
	}
}

func isElement(el xml.StartElement, name string) bool {
	if el.Name.Local == name {
		return true
	}
	if el.Name.Space != "" {
		parts := strings.Split(el.Name.Space, ":")
		if len(parts) > 0 && parts[len(parts)-1] == name {
			return true
		}
	}
	return false
}
