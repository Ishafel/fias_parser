package xmlstream

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamElements_MissingRequiredAttribute(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<ROOT>
        <ITEM ID="1" NAME="Valid" />
        <ITEM NAME="MissingID" />
</ROOT>`

	dir := t.TempDir()
	path := filepath.Join(dir, "data.xml")
	if err := os.WriteFile(path, []byte(xmlData), 0o644); err != nil {
		t.Fatalf("write xml: %v", err)
	}

	var buf bytes.Buffer
	result, err := StreamElements(path, "ITEM", 2, []string{"ID", "NAME"}, &buf)
	if err != nil {
		t.Fatalf("StreamElements() error = %v", err)
	}

	if result.Processed != 1 {
		t.Fatalf("Processed = %d, want 1", result.Processed)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("Skipped = %d, want 1", len(result.Skipped))
	}

	if result.Skipped[0].Error == "" {
		t.Fatalf("expected error message for skipped record, got empty string")
	}

	output := buf.String()
	if output == "" {
		t.Fatalf("expected processed record to be written to buffer")
	}
}
