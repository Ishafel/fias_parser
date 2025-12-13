package xmlstream

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"
)

// Record represents a flattened XML element suitable for JSON serialization.
type Record struct {
	Element    string            `json:"element"`
	Attributes map[string]string `json:"attributes"`
	Content    string            `json:"content,omitempty"`
}

// DetectRoot determines the root element name of an XML file.
func DetectRoot(path string) (string, error) {
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

// StreamElements scans the XML file and serializes each target element into JSON.
func StreamElements(path string, elementName string, out io.Writer) error {
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

func buildRecord(dec *xml.Decoder, start xml.StartElement) (Record, error) {
	attrs := make(map[string]string, len(start.Attr))
	for _, attr := range start.Attr {
		attrs[attr.Name.Local] = attr.Value
	}

	var content strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return Record{}, err
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
				return Record{
					Element:    start.Name.Local,
					Attributes: attrs,
					Content:    content.String(),
				}, nil
			}
		case xml.StartElement:
			if err := dec.Skip(); err != nil {
				return Record{}, err
			}
		}
	}
}
