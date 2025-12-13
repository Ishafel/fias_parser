package schema

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SchemaInfo describes minimal information about an XSD schema.
type SchemaInfo struct {
	Path        string
	RootElement string
}

// LoadSchemas scans the directory for XSD files and builds a map from root element name to schema info.
func LoadSchemas(dir string) (map[string]SchemaInfo, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.xsd"))
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]SchemaInfo)
	for _, path := range matches {
		root, err := DetectXSDRoot(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		schemas[root] = SchemaInfo{Path: path, RootElement: root}
	}

	if len(schemas) == 0 {
		return nil, errors.New("no schemas found")
	}

	return schemas, nil
}

// DetectXSDRoot reads an XSD schema and returns the name attribute of the first <element> tag.
func DetectXSDRoot(path string) (string, error) {
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

func isElement(el xml.StartElement, name string) bool {
	if el.Name.Local == name {
		return true
	}
	if el.Name.Space != "" {
		if parts := strings.Split(el.Name.Space, ":"); len(parts) > 0 && parts[len(parts)-1] == name {
			return true
		}
	}
	return false
}
