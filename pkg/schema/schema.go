package schema

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SchemaInfo describes minimal information about an XSD schema.
type SchemaInfo struct {
	Path        string
	RootElement string
	// RequiredAttributes lists attribute names marked as use="required" for each element defined in the schema.
	RequiredAttributes map[string][]string
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
		required, err := ExtractRequiredAttributes(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		schemas[root] = SchemaInfo{Path: path, RootElement: root, RequiredAttributes: required}
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
	return isElementName(el.Name, name)
}

func isElementName(el xml.Name, name string) bool {
	if el.Local == name {
		return true
	}
	if el.Space != "" {
		if parts := strings.Split(el.Space, ":"); len(parts) > 0 && parts[len(parts)-1] == name {
			return true
		}
	}
	return false
}

// ExtractRequiredAttributes parses the schema and returns a map from element name to the list of attributes
// marked with use="required".
func ExtractRequiredAttributes(path string) (map[string][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(bufio.NewReader(f))

	type elementContext struct {
		name   string
		attrs  map[string]struct{}
		active bool
	}

	var stack []elementContext
	required := make(map[string][]string)

	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if isElement(t, "element") {
				name := ""
				for _, attr := range t.Attr {
					if attr.Name.Local == "name" {
						name = attr.Value
						break
					}
				}
				ctx := elementContext{name: name, attrs: make(map[string]struct{}), active: name != ""}
				stack = append(stack, ctx)
				continue
			}

			if isElement(t, "attribute") && len(stack) > 0 {
				current := &stack[len(stack)-1]
				if !current.active {
					continue
				}
				attrName := ""
				useRequired := false
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "name":
						attrName = attr.Value
					case "use":
						if attr.Value == "required" {
							useRequired = true
						}
					}
				}
				if useRequired && attrName != "" {
					current.attrs[attrName] = struct{}{}
				}
				continue
			}

		case xml.EndElement:
			if isElementName(t.Name, "element") && len(stack) > 0 {
				ctx := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				if ctx.active && len(ctx.attrs) > 0 {
					names := make([]string, 0, len(ctx.attrs))
					for name := range ctx.attrs {
						names = append(names, name)
					}
					sort.Strings(names)
					required[ctx.name] = names
				}
			}
		}
	}

	return required, nil
}
