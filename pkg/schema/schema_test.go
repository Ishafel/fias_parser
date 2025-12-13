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

func TestDatasetPrefix(t *testing.T) {
	tests := map[string]string{
		"AS_ADDR_OBJ_2_251_01_04_01_01.xsd":             "AS_ADDR_OBJ",
		"AS_CHANGE_HISTORY_251_21_04_01_01.xsd":         "AS_CHANGE_HISTORY",
		"AS_ADDR_OBJ_20251204_0ea5af00-5ea6.XML":        "AS_ADDR_OBJ",
		"AS_HOUSES_20240101_guid.xml":                   "AS_HOUSES",
		"SIMPLE_NAME_WITHOUT_VERSION.xsd":               "SIMPLE_NAME_WITHOUT_VERSION",
		"nested/path/AS_APARTMENTS_PARAMS_2_251.xsd":    "AS_APARTMENTS_PARAMS",
		"nested/path/AS_APARTMENTS_PARAMS_20250101.xml": "AS_APARTMENTS_PARAMS",
	}

	for input, expected := range tests {
		if got := DatasetPrefix(input); got != expected {
			t.Fatalf("DatasetPrefix(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestLoadSchemasDuplicatePrefix(t *testing.T) {
	dir := t.TempDir()

	schemaBody := `<?xml version="1.0"?><xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="ITEMS"></xs:element></xs:schema>`

	first := filepath.Join(dir, "AS_ITEMS_1.xsd")
	if err := os.WriteFile(first, []byte(schemaBody), 0o644); err != nil {
		t.Fatalf("write first schema: %v", err)
	}

	second := filepath.Join(dir, "AS_ITEMS_2.xsd")
	if err := os.WriteFile(second, []byte(schemaBody), 0o644); err != nil {
		t.Fatalf("write second schema: %v", err)
	}

	if _, err := LoadSchemas(dir); err == nil {
		t.Fatalf("expected error for duplicate schema prefix, got nil")
	}
}
