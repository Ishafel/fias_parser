package schema_test

import (
	"path/filepath"
	"testing"

	"fias_parser/pkg/schema"
)

func TestDetectXSDRoot(t *testing.T) {
	root, err := schema.DetectXSDRoot(filepath.Join("..", "..", "gar_schemas", "AS_HOUSES_2_251_08_04_01_01.xsd"))
	if err != nil {
		t.Fatalf("detectXSDRoot returned error: %v", err)
	}
	if root != "HOUSES" {
		t.Fatalf("expected root HOUSES, got %s", root)
	}
}
