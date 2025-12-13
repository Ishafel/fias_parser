package xmlstream

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Record represents a flattened XML element serialized to JSON.
type Record struct {
	Element    string            `json:"element"`
	Attributes map[string]string `json:"attributes"`
	Content    string            `json:"content,omitempty"`
}

// DetectXMLRoot returns the name of the XML document root element.
func DetectXMLRoot(path string) (string, error) {
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

// CountElements determines the target element name (if empty) and counts
// how many matching elements exist in the XML document.
func CountElements(path string, elementName string) (string, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	dec := xml.NewDecoder(bufio.NewReader(f))

	depth := 0
	target := elementName
	count := 0

	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return target, count, nil
			}
			return target, count, err
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
				count++
				if err := dec.Skip(); err != nil {
					if errors.Is(err, io.EOF) {
						return target, count, nil
					}
					return target, count, err
				}
				depth--
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
			}
		}
	}
}

// StreamResult captures statistics and skipped records during streaming.
type StreamResult struct {
	Expected  int
	Processed int
	Skipped   []SkippedRecord
}

// SkippedRecord describes a record that could not be processed.
type SkippedRecord struct {
	Index      int
	ByteOffset int64
	Element    string
	Error      string
}

// StreamElements scans an XML file and emits each matching element as a JSON object to the writer.
// It returns a StreamResult with counts and skipped record information.
func StreamElements(path string, elementName string, expected int, out io.Writer) (StreamResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return StreamResult{}, err
	}
	defer f.Close()

	enc := json.NewEncoder(out)
	dec := xml.NewDecoder(bufio.NewReader(f))

	depth := 0
	target := elementName

	processed := 0
	index := 0
	skipped := make([]SkippedRecord, 0)

	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return StreamResult{Expected: expected, Processed: processed, Skipped: skipped}, nil
			}
			return StreamResult{Expected: expected, Processed: processed, Skipped: skipped}, err
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
				index++
				offset := dec.InputOffset()
				rec, err := buildRecord(dec, t)
				if err != nil {
					skipped = append(skipped, SkippedRecord{
						Index:      index,
						ByteOffset: offset,
						Element:    t.Name.Local,
						Error:      err.Error(),
					})
					if skipErr := dec.Skip(); skipErr != nil && !errors.Is(skipErr, io.EOF) {
						return StreamResult{Expected: expected, Processed: processed, Skipped: skipped}, fmt.Errorf("skip element after error: %w", skipErr)
					}
					depth--
					continue
				}
				if err := enc.Encode(rec); err != nil {
					return StreamResult{Expected: expected, Processed: processed, Skipped: skipped}, err
				}
				processed++
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
