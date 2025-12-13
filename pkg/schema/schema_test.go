package schema

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractRequiredAttributes(t *testing.T) {
	xsd := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
        <xs:element name="ROOT">
                <xs:complexType>
                        <xs:sequence>
                                <xs:element name="ITEM">
                                        <xs:complexType>
                                                <xs:attribute name="ID" use="required" />
                                                <xs:attribute name="NAME" use="required" />
                                                <xs:attribute name="OPTIONAL" use="optional" />
                                        </xs:complexType>
                                </xs:element>
                        </xs:sequence>
                </xs:complexType>
        </xs:element>
</xs:schema>`

	dir := t.TempDir()
	path := filepath.Join(dir, "schema.xsd")
	if err := os.WriteFile(path, []byte(xsd), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	attrs, err := ExtractRequiredAttributes(path)
	if err != nil {
		t.Fatalf("ExtractRequiredAttributes() error = %v", err)
	}

	expected := map[string][]string{
		"ITEM": []string{"ID", "NAME"},
	}

	if !reflect.DeepEqual(attrs, expected) {
		t.Fatalf("ExtractRequiredAttributes() = %#v, want %#v", attrs, expected)
	}
}
