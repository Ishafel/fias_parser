package schemas

import (
	"path/filepath"
	"testing"
)

func TestDetectRoot(t *testing.T) {
	root, err := DetectRoot(filepath.Join("..", "..", "gar_schemas", "AS_HOUSES_2_251_08_04_01_01.xsd"))
	if err != nil {
		t.Fatalf("DetectRoot returned error: %v", err)
	}
	if root != "HOUSES" {
		t.Fatalf("expected root HOUSES, got %s", root)
	}
}
